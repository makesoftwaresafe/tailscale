package main

import (
	"flag"
	"log"
	"net/http"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

var (
	dev = flag.Bool("dev", true, "Run in dev mode (watching JS and CSS)")
)

func main() {
	flag.Parse()

	buildOptions := esbuild.BuildOptions{
		EntryPoints: []string{"index.js", "index.css"},
		Outdir:      "./dist",
		Bundle:      true,
		Sourcemap:   esbuild.SourceMapLinked,
		LogLevel:    esbuild.LogLevelInfo,
	}

	if *dev {
		result, err := esbuild.Serve(esbuild.ServeOptions{
			Port:     9090,
			Servedir: "./",
		}, buildOptions)
		if err != nil {
			log.Fatal(err)
			return
		}
		log.Printf("Listening on http://%s:%d\n", result.Host, result.Port)
		result.Wait()
	} else {
		buildOptions.Write = true
		buildOptions.MinifyWhitespace = true
		buildOptions.MinifyIdentifiers = true
		buildOptions.MinifySyntax = true

		result := esbuild.Build(buildOptions)
		if len(result.Errors) > 0 {
			log.Printf("ESBuild Error:\n")
			for _, e := range result.Errors {
				log.Printf("%v", e)
			}
			log.Fatal("Build failed")
		}
		if len(result.Warnings) > 0 {
			log.Printf("ESBuild Warnings:\n")
			for _, w := range result.Warnings {
				log.Printf("%v", w)
			}
		}
		log.Printf("Listening on http://localhost:9090")
		err := http.ListenAndServe(":9090", http.FileServer(http.Dir(".")))
		if err != nil {
			log.Fatal(err)
		}
	}
}
