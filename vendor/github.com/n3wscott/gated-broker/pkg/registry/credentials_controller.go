package registry

import (
	"fmt"

	"github.com/pborman/uuid"
)

func (c *ControllerInstance) AssignCredentials(osbInstanceId OsbId, osbBindingId OsbId) (*LightBinding, error) {
	// find the light id from the osb instance id
	lightId := c.OsbInstanceIdToId[osbInstanceId]
	if lightId == "" {
		return nil, fmt.Errorf("error: no light registered for OsbId[%s]", osbInstanceId)
	}

	// find the instance from the light id
	instance := c.IdToInstance[lightId]
	if instance == nil {
		return nil, fmt.Errorf("error: no light registered for OsbId[%s] and registery is in a bad state", osbInstanceId)
	}

	// make sure the binding id is not tied to a light
	if c.OsbBindingIdToId[osbBindingId] != "" {
		return nil, fmt.Errorf("error: bindingId[%s] in use", osbBindingId)
	}

	// make sure the instance does not have a conflicting binding
	for _, binding := range instance.Bindings {
		if binding.OsbBindingId == osbBindingId {
			return nil, fmt.Errorf("error: bindingId[%s] in use and we are in a bad state", osbBindingId)
		}
	}

	// ok to make a binding for the given instance
	binding := LightBinding{
		OsbBindingId: osbBindingId,
		Id:           lightId,
		Secret:       Secret(uuid.NewUUID().String()),
	}
	// TODO: confirm that the secret is unique?

	c.OsbBindingIdToId[osbBindingId] = lightId
	c.SecretToId[binding.Secret] = lightId
	instance.Bindings = append(instance.Bindings, binding)

	return &binding, nil
}

func (c *ControllerInstance) RemoveCredentials(osbBindingId OsbId) error {

	// get the light id from binding id
	lightId := c.OsbBindingIdToId[osbBindingId]
	if lightId == "" {
		return fmt.Errorf("error: bindingId[%s] is not in use", osbBindingId)
	}

	// get the instance from light id
	instance := c.IdToInstance[lightId]

	// remove the binding from the instance
	var binding *LightBinding
	for i, b := range instance.Bindings {
		if b.OsbBindingId == osbBindingId {
			binding = &b
			// swap the binding to the end of the list and decrease the list
			// TODO: it could be simpler to just use a map for instance bindings
			instance.Bindings[i] = instance.Bindings[len(instance.Bindings)-1]
			instance.Bindings = instance.Bindings[:len(instance.Bindings)-1]
			break
		}
	}

	if binding == nil {
		return fmt.Errorf("error: bindingId[%s] is mapped but not part of the instance[%s]", osbBindingId, instance.OsbInstanceId)
	}

	// clean up the maps
	c.SecretToId[binding.Secret] = ""
	c.OsbBindingIdToId[osbBindingId] = ""

	return nil
}
