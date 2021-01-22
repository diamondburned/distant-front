package index

import (
	"testing"

	"github.com/diamondburned/distant-front/lib/distance"
	"github.com/hexops/autogold"
)

var sortInput = []distance.Player{
	newPlayer(false, -1, true, false),  // alive
	newPlayer(false, -1, false, false), // dead
	newPlayer(false, 1, true, false),   // alive
	newPlayer(false, -1, false, true),  // spectating
	newPlayer(false, -1, false, true),  // spectating
	newPlayer(true, 1, true, false),
	newPlayer(true, 3, true, false),
	newPlayer(true, 4, true, false),
	newPlayer(true, 2, true, false),
}

func TestSort(t *testing.T) {
	want := autogold.Want("sort", []distance.Player{
		newPlayer(true, 1, true, false),
		newPlayer(true, 2, true, false),
		newPlayer(true, 3, true, false),
		newPlayer(true, 4, true, false),
		newPlayer(false, 1, true, false),   // alive
		newPlayer(false, -1, false, false), // dead
		newPlayer(false, -1, true, false),  // alive
		newPlayer(false, -1, false, true),  // spectating
		newPlayer(false, -1, false, true),  // spectating
	})

	want.Equal(t, sortPlayers(sortInput))
}

func BenchmarkSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sortPlayers(sortInput)
	}
}

func newPlayer(finished bool, finishTime int, alive, spec bool) distance.Player {
	return distance.Player{
		Car: distance.Car{
			Finished:   finished,
			FinishData: finishTime,
			Alive:      alive,
			Spectator:  spec,
		},
	}
}
