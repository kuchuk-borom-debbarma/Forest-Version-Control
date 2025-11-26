package versioncontrol

func Commit(message string, author string, files []string) {
	/*
		Calculate hash and store all files as objects
		Then for the directories create tree where entries will be other sub directories and files
		Hash the trees and save them as well in objects
		Create and save commit object and make it point to the root directory tree
		Set the head to point to the commit hash
	*/
}
