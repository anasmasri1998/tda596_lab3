package main

type SuccFind struct {
	Id string
}

type Bingo struct {
	Identified bool
	SuccId     NodeIdentifier
}

type Get struct {
	Id string
}

type Get_reply struct {
	Content string
	Confirm bool
}

type Put struct {
	Id    string
	Value string
}

type Put_reply struct {
	Confirm bool
}

type Delete struct {
	Id string
}

type Delete_reply struct {
	Confirm bool
}

type Empty struct {
	confirm bool
}

// type Bucket struct {
// 	Id    string
// 	Value string
// }

// type Bucket_reply struct {
// 	confirm bool
// 	Content string
// }
