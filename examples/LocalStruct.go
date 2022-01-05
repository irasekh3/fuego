/*
 * Created by Ilan Rasekh on 2019/9/24
 * Copyright (c) 2019. All rights reserved.
 */

package main

import "github.com/irasekh3/fuego"

type MyMath struct {
	Offset float64
}

func (m MyMath) Add(a float64, b float64) float64 {
	return a + b + m.Offset
}

func (m MyMath) Subtract(a float64, b float64) float64 {
	return a - b - m.Offset
}

func main() {
	fuego.Fuego(&MyMath{})
}
