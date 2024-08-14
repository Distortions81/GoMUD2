package main

type objectData struct {
	ObjType     int
	Name        string
	Description string
	VNUM        int
	UUID        uuidData

	Contents []objectData
}

type mobData struct {
	MobType     int
	Name        string
	Description string
	VNUM        int
	UUID        uuidData

	Contents []objectData
}

type resetData struct {
	ResetType   int
	Name        string
	Description string
	VNUM        int
	UUID        uuidData
}
