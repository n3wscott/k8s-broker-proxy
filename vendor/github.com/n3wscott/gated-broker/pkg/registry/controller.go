package registry

func NewControllerInstance(lights map[Location]map[Kind]int) *ControllerInstance {
	c := ControllerInstance{}

	c.populateLightInstancesFromLights(lights)

	return &c
}
