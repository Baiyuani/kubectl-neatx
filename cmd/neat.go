package cmd

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Neat gets a Kubernetes resource json as string and de-clutters it to make it more readable.
func Neat(in string) (string, error) {
	var err error
	draft := in
	kind := gjson.Get(in, "kind").String()

	if in == "" {
		return draft, fmt.Errorf("error in neatPod, input json is empty")
	}
	if !gjson.Valid(in) {
		return draft, fmt.Errorf("error in neatPod, input is not a vaild json: %s", in[:20])
	}

	// handle list
	if kind == "List" {
		items := gjson.Get(draft, "items").Array()
		for i, item := range items {
			itemNeat, err := Neat(item.String())
			if err != nil {
				continue
			}
			draft, err = sjson.SetRaw(draft, fmt.Sprintf("items.%d", i), itemNeat)
			if err != nil {
				continue
			}
		}
		// general neating
		draft, err = neatMetadata(draft, kind)
		if err != nil {
			return draft, fmt.Errorf("error in neatMetadata : %v", err)
		}
		return draft, nil
	}

	// defaults neating
	// draft, err = defaults.NeatDefaults(draft)
	// if err != nil {
	// 	return draft, fmt.Errorf("error in neatDefaults : %v", err)
	// }

	draft, err = neatSpec(draft, kind)
	if err != nil {
		return draft, fmt.Errorf("error in neatSpec : %v", err)
	}

	// controllers neating
	// draft, err = neatScheduler(draft)
	// if err != nil {
	// 	return draft, fmt.Errorf("error in neatScheduler : %v", err)
	// }
	if kind == "Pod" {
		draft, err = neatServiceAccount(draft)
		if err != nil {
			return draft, fmt.Errorf("error in neatServiceAccount : %v", err)
		}
	}

	// general neating
	draft, err = neatMetadata(draft, kind)
	if err != nil {
		return draft, fmt.Errorf("error in neatMetadata : %v", err)
	}
	draft, err = neatStatus(draft)
	if err != nil {
		return draft, fmt.Errorf("error in neatStatus : %v", err)
	}
	// draft, err = neatEmpty(draft)
	// if err != nil {
	// 	return draft, fmt.Errorf("error in neatEmpty : %v", err)
	// }

	return draft, nil
}

func neatSpec(in string, kind string) (string, error) {
	var draft string
	// var err   error

	draft = in

	draft, _ = sjson.Delete(draft, "spec.template.metadata.creationTimestamp")
	// if err != nil {
	// 	return draft, fmt.Errorf("error deleting spec.template.metadata.creationTimestamp : %v", err)
	// }
	if kind == "Service" {
		draft, _ = sjson.Delete(draft, "spec.clusterIP")
		draft, _ = sjson.Delete(draft, "spec.clusterIPs")
		// if err != nil {
		// 	return draft, fmt.Errorf("error deleting spec.clusterIP : %v", err)
		// }
	}
	if kind == "PersistentVolume" {
		draft, _ = sjson.Delete(draft, "spec.claimRef")
	}
	if kind == "PersistentVolumeClaim" {
		draft, _ = sjson.Delete(draft, `metadata.annotations.pv\.kubernetes\.io/bound-by-controller`)
		draft, _ = sjson.Delete(draft, `metadata.annotations.pv\.kubernetes\.io/bind-completed`)
	}
	if kind == "Deployment" {
		draft, _ = sjson.Delete(draft, `spec.template.metadata.annotations.kubectl\.kubernetes\.io/restartedAt`)
	}

	return draft, nil
}

func neatMetadata(in string, kind string) (string, error) {
	var err error

	in, _ = sjson.Delete(in, `metadata.annotations.kubectl\.kubernetes\.io/last-applied-configuration`)
	if kind == "Deployment" {
		in, _ = sjson.Delete(in, `metadata.annotations.deployment\.kubernetes\.io/revision`)
	}
	// if err != nil {
	// 	return in, fmt.Errorf("error deleting last-applied-configuration : %v", err)
	// }
	// TODO: prettify this. gjson's @pretty is ok but setRaw the pretty code gives unwanted result
	newMeta := gjson.Get(in, "{metadata.name,metadata.namespace,metadata.labels,metadata.annotations}")
	in, err = sjson.Set(in, "metadata", newMeta.Value())
	if err != nil {
		return in, fmt.Errorf("error setting new metadata : %v", err)
	}
	return in, nil
}

func neatStatus(in string) (string, error) {
	return sjson.Delete(in, "status")
}

func neatScheduler(in string) (string, error) {
	return sjson.Delete(in, "spec.nodeName")
}

func neatServiceAccount(in string) (string, error) {
	// var err error
	// keep an eye open on https://github.com/tidwall/sjson/issues/11
	// when it's implemented, we can do:
	// sjson.delete(in, "spec.volumes.#(name%default-token-*)")
	// sjson.delete(in, "spec.containers.#.volumeMounts.#(name%default-token-*)")

	// for vi, v := range gjson.Get(in, "spec.volumes").Array() {
	// 	vname := v.Get("name").String()
	// 	if strings.HasPrefix(vname, "default-token-") {
	// 		in, err = sjson.Delete(in, fmt.Sprintf("spec.volumes.%d", vi))
	// 		if err != nil {
	// 			continue
	// 		}
	// 	}
	// }
	// for ci, c := range gjson.Get(in, "spec.containers").Array() {
	// 	for vmi, vm := range c.Get("volumeMounts").Array() {
	// 		vmname := vm.Get("name").String()
	// 		if strings.HasPrefix(vmname, "default-token-") {
	// 			in, err = sjson.Delete(in, fmt.Sprintf("spec.containers.%d.volumeMounts.%d", ci, vmi))
	// 			if err != nil {
	// 				continue
	// 			}
	// 		}
	// 	}
	// }
	in, _ = sjson.Delete(in, "spec.serviceAccount") //Deprecated: Use serviceAccountName instead

	return in, nil
}

// neatEmpty removes all zero length elements in the json
func neatEmpty(in string) (string, error) {
	var err error
	jsonResult := gjson.Parse(in)
	var empties []string
	findEmptyPathsRecursive(jsonResult, "", &empties)
	for _, emptyPath := range empties {
		// if we just delete emptyPath, it may create empty parents
		// so we walk the path and re-check for emptiness at every level
		emptyPathParts := strings.Split(emptyPath, ".")
		for i := len(emptyPathParts); i > 0; i-- {
			curPath := strings.Join(emptyPathParts[:i], ".")
			cur := gjson.Get(in, curPath)
			if isResultEmpty(cur) {
				in, err = sjson.Delete(in, curPath)
				if err != nil {
					continue
				}
			}
		}
	}
	return in, nil
}

// findEmptyPathsRecursive builds a list of paths that point to zero length elements
// cur is the current element to look at
// path is the path to cur
// res is a pointer to a list of empty paths to populate
func findEmptyPathsRecursive(cur gjson.Result, path string, res *[]string) {
	if isResultEmpty(cur) {
		*res = append(*res, path[1:]) //remove '.' from start
		return
	}
	if !(cur.IsArray() || cur.IsObject()) {
		return
	}
	// sjson's ForEach doesn't put track index when iterating arrays, hence the index variable
	index := -1
	cur.ForEach(func(k gjson.Result, v gjson.Result) bool {
		var newPath string
		if cur.IsArray() {
			index++
			newPath = fmt.Sprintf("%s.%d", path, index)
		} else {
			newPath = fmt.Sprintf("%s.%s", path, k.Str)
		}
		findEmptyPathsRecursive(v, newPath, res)
		return true
	})
}

func isResultEmpty(j gjson.Result) bool {
	v := j.Value()
	switch vt := v.(type) {
	// empty string != lack of string. keep empty strings as it's meaningful data
	// case string:
	// 	return vt == ""
	case []interface{}:
		return len(vt) == 0
	case map[string]interface{}:
		return len(vt) == 0
	}
	return false
}
