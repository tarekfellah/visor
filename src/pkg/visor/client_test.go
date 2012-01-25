package visor

import ()

func clientSetup() (c *Client) {
	c = createClient()
	c.Deldir("/apps", c.Rev)

	return
}

// HELPER
func createClient() (c *Client) {
	c, err := Dial(DEFAULT_ADDR)
	if err != nil {
		panic(err)
	}

	return
}
