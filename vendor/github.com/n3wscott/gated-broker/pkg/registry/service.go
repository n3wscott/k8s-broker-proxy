package registry

import (
	"bytes"

	"fmt"

	"github.com/pborman/uuid"
)

func (c *ControllerInstance) populateLightInstancesFromLights(lights map[Location]map[Kind]int) {

	//glog.Infof("Going to process %d locations.", len(lights))

	c.IdToLight = make(map[LightId]*Light)
	c.IdToInstance = make(map[LightId]*LightInstance)
	c.LocationKindToIds = make(map[Location]map[Kind][]LightId)
	c.OsbInstanceIdToId = make(map[OsbId]LightId)
	c.OsbBindingIdToId = make(map[OsbId]LightId)
	c.SecretToId = make(map[Secret]LightId)

	for location, kinds := range lights {
		//		glog.Infof("For location %s, %d kinds", location, len(kinds))

		c.LocationKindToIds[location] = make(map[Kind][]LightId)

		for kind, count := range kinds {
			//			glog.Infof("\tFor kind %s, %d lights", kind, count)

			for i := 0; i < count; i++ {
				light := newLight(location, kind)
				c.IdToLight[light.Id] = &light
				c.LocationKindToIds[location][kind] = append(c.LocationKindToIds[location][kind], light.Id)
			}
		}
	}
}

func (c *ControllerInstance) lightIsReserved(lightId LightId) bool {
	if c.IdToInstance[lightId] != nil {
		return true
	}
	return false
}

func newLight(location Location, kind Kind) Light {
	lightId := LightId(uuid.NewUUID().String())
	light := Light{
		Id:       lightId,
		Location: location,
		Kind:     kind,
		//  TODO: need to init for the type of light this is.
	}
	return light
}

func (c *ControllerInstance) String() string {
	var buffer bytes.Buffer

	//buffer.WriteString("IdToInstance:\n")
	//for lightId, light := range c.IdToLight {
	//buffer.WriteString(fmt.Sprintf("\t%s:%s,\n", lightId, light.String()))
	//}

	buffer.WriteString("Location/Kinds:\n")

	for location, kinds := range c.LocationKindToIds {
		buffer.WriteString(fmt.Sprintf("\t%s,\n", location))
		for kind, lights := range kinds {
			available := 0
			lightIntensities := ""
			for _, lightId := range lights {
				if !c.lightIsReserved(lightId) {
					available++
				}
				light := c.IdToLight[lightId]
				lightIntensities += fmt.Sprintf("%.2f ", light.Intensity)
			}
			buffer.WriteString(fmt.Sprintf("\t\t%s: %d of %d available [%s]\n", kind, available, len(lights), lightIntensities))
		}
	}

	return buffer.String()
}

func (l *Light) String() string {
	return fmt.Sprintf("[%s] at %s", l.Kind, l.Location)
}
