package main

import (
	"time"

	do "gopkg.in/godo.v2"
)

func tasks(p *do.Project) {
	watch := []string{
		"ui/package.json",
		"ui/bower.json",
		"ui/.bowerrc",
		"ui/webpack.config.js",
		"ui/app/js/**/*.js",
		"ui/app/sass/*.scss",
		"ui/public/index.html",
		"**/*.go",
		"!apiserver/bindata_assetfs.go",
	}

	p.Task("default", do.S{"build"}, nil)

	p.Task("build", nil, func(c *do.Context) {
		c.Run("make build")
	}).Src(watch...).Debounce(3 * time.Second)

	p.Task("start", do.S{"build"}, func(c *do.Context) {
		c.Start("./blamedns")
	})
}

func main() {
	do.Godo(tasks)
}
