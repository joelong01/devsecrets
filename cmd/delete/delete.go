package delete

import (

	"devsecrets/globals"
)

/*
a quick but possibly unsafe delete -- delete the specified repo in GitHub and delete the ResourceGroup
*/
func onDelete() {
	globals.EchoInfo(globals.ClearLineRight + "Deleting Portal Config\n")
	
}
