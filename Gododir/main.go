package main

import do "gopkg.in/godo.v2"

func tasks(p *do.Project) {
	p.Task("default", do.S{"build"}, nil)

	p.Task("npm-install", nil, func(c *do.Context) {
		c.Run("make npm-install")
	}).Src("ui/package.json")

	p.Task("bower-install", nil, func(c *do.Context) {
		c.Run("make bower-install")
	}).Src("ui/bower.json", "ui/.bowerrc")

	p.Task("webpack", do.S{"npm-install", "bower-install"}, func(c *do.Context) {
		c.Run("make webpack")
	}).Src("ui/webpack.config.js", "ui/app/jsx/**/*.jsx", "ui/app/sass/*.scss")

	p.Task("generate", do.S{"webpack"}, func(c *do.Context) {
		c.Run("make generate")
	}).Src(".webpack-stamp", "ui/public/index.html")

	p.Task("build", do.S{"generate"}, func(c *do.Context) {
		c.Run("make build")
	}).Src("**/*.go")

	p.Task("start", do.S{"build"}, func(c *do.Context) {
		c.Start("./blamedns")
	}).Src("blamedns").Debounce(3000)
}

func main() {
	do.Godo(tasks)
}
