// +build ignore

package main

import (
	"github.com/iancoleman/strcase"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

const templateSource = `// Generated by events_gen.go

package events

import (
	"github.com/jonas747/discordgo"
)

const ( {{range $k, $v := .}}
	Event{{.Name}} = "{{.ProperName}}"{{end}}
)

var AllEvents = []string{ {{range .}}{{if .Discord}}
	Event{{.Name}},{{end}}{{end}}
}

var NewEvents = map[string]func() interface{}{ {{range .}}{{if .Discord}}
	"{{.ProperName}}": func() interface{} {
		return &discordgo.{{.Name}}{}
	},{{end}}{{end}}
}

{{range .}}{{if .Discord}}
func (event *EventData) {{.Name}}() *discordgo.{{.Name}}{
	return event.Event.(*discordgo.{{.Name}})
}
{{end}}{{end}}`

type Event struct {
	Name       string
	ProperName string
	Discord    bool
}

var (
	parsedTemplate = template.Must(template.New("").Parse(templateSource))
	outputFile     = "../all_events.go"
)

func main() {
	pkg, err := packages.Load(&packages.Config{}, "github.com/jonas747/discordgo")
	if err != nil {
		panic(err)
	}

	path := ""
	for _, p := range pkg[0].GoFiles {
		if strings.HasSuffix(p, "/events.go") {
			path = p
			break
		}
	}

	log.Println("Found path: ", path)

	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, path, nil, 0)
	if err != nil {
		log.Fatalf("Failed to parse events.go: %s", err)
		return
	}

	var names []string
	for name := range parsedFile.Scope.Objects {
		names = append(names, name)
	}

	sort.Strings(names)

	events := make([]Event, len(names))
	events[0] = Event{Name: "All", ProperName: "all", Discord: false}

	i := 1
	for _, name := range names {
		if name == "Event" {
			continue
		}

		events[i] = Event{
			Name:       name,
			ProperName: strcase.ToSnake(name),
			Discord:    true,
		}

		i++
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}

	err = parsedTemplate.Execute(file, events)
	if err != nil {
		log.Fatalf("Failed to execute template: %s", err)
	}

	err = file.Close()
	if err != nil {
		log.Fatalf("Failed to save and close file: %s", err)
	}

	log.Println("The file is generated successfully")
}
