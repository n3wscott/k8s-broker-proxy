package registry

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"image"
	"image/jpeg"
	"net/http"
	"os/exec"
	"strconv"

	"fmt"

	"github.com/golang/glog"
	"github.com/tmc/dot"
)

// this needs dot (graphviz) to be installed locally to the server
func (c *ControllerInstance) HandleGetGraph(w http.ResponseWriter, r *http.Request) {
	dot := c.Graph()
	writeImageFromDot(w, dot)
}

var ImageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><img src="data:image/jpg;base64,{{.Image}}"></body>`

func writeImageWithTemplate(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		glog.Error("unable to encode image.")
	}

	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	if tmpl, err := template.New("image").Parse(ImageTemplate); err != nil {
		glog.Error("unable to parse image template.")
	} else {
		data := map[string]interface{}{"Image": str}
		if err = tmpl.Execute(w, data); err != nil {
			glog.Error("unable to execute template.")
		}
	}
}

func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		glog.Error("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		glog.Error("unable to write image.")
	}
}

func writeImageFromDot(w http.ResponseWriter, dot string) {
	var png bytes.Buffer

	dotBuffer := bytes.NewBufferString(dot)

	cmd := exec.Command("dot", "-Tpng")
	cmd.Stdin = dotBuffer
	cmd.Stdout = &png
	err := cmd.Run()
	if err != nil {
		glog.Error(err)
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(png.Bytes())))
	if _, err := w.Write(png.Bytes()); err != nil {
		glog.Error("unable to write image.")
	}
}

func (c *ControllerInstance) Graph() string {

	g := dot.NewGraph("G")
	g.Set("label", "Light Registry")
	g.Set("rankdir", "LR")

	for location, kinds := range c.LocationKindToIds {

		// Add the location and point the registry to it
		locationNode := dot.NewNode(string(location))
		g.AddNode(locationNode)

		for kind, lights := range kinds {

			kindNode := dot.NewNode(string(location) + string(kind))
			kindNode.Set("label", string(kind))
			g.AddNode(kindNode)
			kindEdge := dot.NewEdge(locationNode, kindNode)
			kindEdge.Set("dir", "none")
			g.AddEdge(kindEdge)

			for _, lightId := range lights {
				light := c.IdToLight[lightId]

				lightNode := dot.NewNode(string(lightId))
				color := fmt.Sprintf("grey%d", int(light.Intensity*100))
				glog.Info("color = ", color)
				lightNode.Set("fillcolor", color)
				lightNode.Set("style", "filled")
				textColor := "grey100"
				if light.Intensity > .5 {
					textColor = "grey0"
				}
				lightNode.Set("fontcolor", textColor)
				g.AddNode(lightNode)
				lightEdge := dot.NewEdge(kindNode, lightNode)
				lightEdge.Set("dir", "none")
				g.AddEdge(lightEdge)
			}
		}
	}
	return g.String()
}
