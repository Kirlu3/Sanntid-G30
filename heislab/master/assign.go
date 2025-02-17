package Master

import(
	"fmt"
	"runtime"
	"encoding/json"
	"os/exec"
)

func assign(input HRAInput) map[string][][2]bool{

	hraExecutable  := ""

	switch runtime.GOOS{
        case "linux":   hraExecutable = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }


	// makes input into json format 
	inputJsonFormat, errMarsial := json.Marshal(input)

	if errMarsial != nil{
		fmt.Println("Error using json.Marshal: ", errMarsial)
	}

	// assign and returns output in json format 
	outputJsonFormat, errAssign := exec.Command("./Project-resources/cost_fns/hall_request_assigner/"+hraExecutable, "-i", string(inputJsonFormat)).CombinedOutput()
    
	if errAssign!= nil {
		fmt.Println("Error occured when assigning: ",errAssign)
	}

	output := new(map[string][][2]bool)
	errUnmarshal := json.Unmarshal(outputJsonFormat, &output)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	return *output

}