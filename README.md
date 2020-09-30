Microdata
=========

__HTML Microdata__ is a markup specification often used in combination with the [schema collection][3] to make it easier for search engines to identify and understand content on web pages. One of the most common schema is the rating you see when you google for something. Other schemas are persons, places, events, products, etc.

Microdata is a package for the Go programming language to extract [HTML Microdata][0] from HTML5 documents. It depends on the [golang.org/x/net/html][1] HTML5-compliant parser.

Installation
------------

Build from source:

```sh
$ go get -u github.com/damian-szulc/microdata/cmd/microdata
```


Usage
-----

Parse from URL:

```sh
$ microdata https://www.gog.com/game/...
{
  "items": [
    {
      "type": [
        "http://schema.org/Product"
      ],
      "properties": {
        "additionalProperty": [
          {
            "type": [
              "http://schema.org/PropertyValue"
            ],
{
...
```


Parse HTML from the stdin:

```
$ cat saved.html | microdata
```


Format the output with a Go template to return the "price" property:

```sh
$ microdata -format '{{with index .Items 0}}{{with index .Properties "offers" 0}}{{with index .Properties "price" 0 }}{{ . }}{{end}}{{end}}{{end}}' https://www.gog.com/game/...
8.99
```


Features
--------

- Windows/BSD/Linux supported
- Format output with Go templates
- Parse from Stdin


Go Package
----------

```go
package main

import (
	"encoding/json"
	"os"

	"github.com/damian-szulc/microdata"
)

func main() {
	var data microdata.Microdata
	data, _ = microdata.ParseURL("http://example.com/blogposting")
	b, _ := json.MarshalIndent(data, "", "  ")
	os.Stdout.Write(b)
}
```
