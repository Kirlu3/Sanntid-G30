package master

import(
	"fmt"
	"runtime"
	"os/exec"
)


func assign(state slave.WordlView) map[string][slave.N_FLOORS][2]bool{

	hraExecutable  := ""

	switch runtime.GOOS{
        case "linux":   hraExecutable = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }


	input := transformInput(state) // transforms input from worldview to HRAInput

	// assign and returns output in json format 
	outputJsonFormat, errAssign := exec.Command("./Project-resources/cost_fns/hall_request_assigner/"+hraExecutable, "-i", string(input)).CombinedOutput()
    
	if errAssign!= nil {
		fmt.Println("Error occured when assigning: ",errAssign)
	}

	// transforms output from json format to the ouputformat
	output := transformOutput(outputJsonFormat)

	return output
}

/* TODO:
- ta inn riktig type
- gjøre om fra worldview til HRAInput
- velge riktig output type 
- legge inn i assingOrders funksjonen i main-branchen
*/