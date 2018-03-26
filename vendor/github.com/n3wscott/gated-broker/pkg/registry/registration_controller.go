package registry

import "fmt"

// TODO: update kind to type.

func (c *ControllerInstance) Register(osbInstanceId OsbId, location Location, kind Kind) (*LightInstance, error) {
	// Make sure the osb id is not in use already.
	if c.OsbInstanceIdToId[osbInstanceId] != "" {
		return nil, fmt.Errorf("error: Instance[%s] already registered", osbInstanceId)
	}

	// find if there is a free light in the location given.
	lights := c.LocationKindToIds[location][kind] // TODO: harden this.
	if len(lights) == 0 {
		return nil, fmt.Errorf("error: no lights for %s/%s found", location, kind)
	}

	var instance *LightInstance

	for _, lightId := range lights {
		if c.IdToInstance[lightId] == nil {
			// assign the LightId and OsbId to a new Light Instance.
			instance = &LightInstance{
				OsbInstanceId: osbInstanceId,
				Id:            lightId,
			}
			c.OsbInstanceIdToId[osbInstanceId] = lightId
			c.IdToInstance[lightId] = instance
			break
		}
	}

	// light := c.IdToLight[lightId]

	return instance, nil
}

func (c *ControllerInstance) Deregister(osbInstanceId OsbId) error {

	// Make sure the osb id is in use already.
	lightId := c.OsbInstanceIdToId[osbInstanceId]
	if lightId == "" {
		return fmt.Errorf("error: Instance[%s] is not registered", osbInstanceId)
	}

	// get the instance
	instance := c.IdToInstance[lightId]

	// confirm there are no bindings for the instance
	if len(instance.Bindings) > 0 {
		return fmt.Errorf("error: Instance[%s] has active bindings", osbInstanceId)
	}

	// remove the instance and disconnect the maps
	c.IdToInstance[lightId] = nil
	c.OsbInstanceIdToId[osbInstanceId] = ""

	// get the light and Default it
	light := c.IdToLight[lightId]
	light.Default()

	return nil
}
