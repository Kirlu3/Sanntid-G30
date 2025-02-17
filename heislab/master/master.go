package Master

import (
	"fmt"
)

type HRAElevState struct {
    Floor       int         `json:"floor"` 
    Behavior    string      `json:"behaviour"`
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
	Obstruction bool		`json:"obstruction"`
}

type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"` // first bool is for up and second is down
    States          map[string]HRAElevState     `json:"states"`
}


func Master() {

	requests := [][2]bool{} // list for hall requests 


	input := HRAInput{
		HallRequests: requests,
		States: map[string]HRAElevState{
			"first": HRAElevState{
				Behavior: "idle",
				Floor: 1,
				Direction: "stop",
				CabRequests: []bool{true, false, false, false},
				Obstruction: false,
			},
			"second": HRAElevState{
				Behavior: "moving",
				Floor: 2,
				Direction: "up",
				CabRequests: []bool{false, false, false, true},
				Obstruction: false,
			},
		},
	}
	

	output := assign(input)

	fmt.Println(output)

}


// elevator.Requests[][elevi.BT_Cab] - cab requestene til spesifikk heis



// HUSK 
// hvordan få tak i alle hall requests 
// gjøre om fra struct til riktig input oppsett 
// obstruction!
// requests listen må oppdateres når slavene sier ifra om en request til master