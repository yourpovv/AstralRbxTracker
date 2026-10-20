package tracker

import "fmt"

// new signals is a new detector
type detector func(previous, current Snapshot) (Change, bool)

var detectors = []detector{nameChanged, descriptionChanged, iconChanged}

func PublicChanges(previous, current Snapshot) []Change {
	changes := []Change{}
	for _, detect := range detectors {
		if change, found := detect(previous, current); found {
			changes = append(changes, change)
		}
	}
	return changes
}

func nameChanged(previous, current Snapshot) (Change, bool) {
	summary := fmt.Sprintf("Name: %q → %q", previous.Name, current.Name)
	return Change{Label: "Name", Summary: summary}, previous.Name != current.Name
}

func descriptionChanged(previous, current Snapshot) (Change, bool) {
	change := Change{Label: "Description", Summary: "Description edited"}
	return change, previous.Description != current.Description
}

func iconChanged(previous, current Snapshot) (Change, bool) {
	change := Change{Label: "Icon", Summary: "Icon changed"}
	return change, previous.IconURL != current.IconURL
}
