package main

type objectData struct {
	ObjType     int
	Name        string
	Description string
	VNUM        int
	UUID        string

	Contents []objectData
}
