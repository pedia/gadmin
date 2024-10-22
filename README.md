# GAdmin
`GAdmin` solves the boring problem of building an admin interface on top of an existing data model. With little effort, it lets you manage your web service’s data through a user-friendly interface.

## How does it work? 
The basic concept behind `GAdmin`, is that it lets you build complicated interfaces by grouping individual views together in classes: Each web page you see on the frontend, represents a method on a class that has explicitly been added to the interface.

We write `GAdmin` in Go, heavily use `net/http`, `template/html` and `gorm`.


# Flask-Admin
`Flask-Admin` inspires this project. We copy the template files from flask-admin(bootstrap4 only).

# Status
[examples/simple](examples/simple/main.go)
