/*
Application Server bootstrapper.

@author Thanh Nguyen <btnguyen2k@gmail.com>
@since template-v0.4.r1
*/
package main

import (
	"main/src/prista"
	"math/rand"
	"time"
)

func main() {
	// it is a good idea to initialize random seed
	rand.Seed(time.Now().UnixNano())
	prista.Start()
}
