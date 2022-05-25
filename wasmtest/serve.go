package main

import (
	"log"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

func main() {

	result, err := esbuild.Serve(esbuild.ServeOptions{
		Port:     9090,
		Servedir: "./",
	}, esbuild.BuildOptions{
		EntryPoints: []string{"index.js", "index.css"},
		Outdir:      "./dist",
		Bundle:      true,
		Sourcemap:   esbuild.SourceMapLinked,
		LogLevel:    esbuild.LogLevelInfo,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("listening on http://%s:%d\n", result.Host, result.Port)
	result.Wait()
}
