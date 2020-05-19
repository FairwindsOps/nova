package output

import (
	"encoding/json"
	"io/ioutil"
	"k8s.io/klog"
)

// FileArgs contain args for the File Output
type FileArgs struct {
	Path string
}

// FileOutput represents a message to be sent to file
type FileOutput struct {
	Output []ReleaseOutput `json:"helm_releases"`
}

// Send dispatches a message to file
func (f *FileOutput) Send(args *FileArgs) error {
	data, err := json.Marshal(f)
	if err != nil {
		klog.Errorf("Error marshaling json: %v", err)
		return err
	}

	err = ioutil.WriteFile(args.Path, data, 0644)
	if err != nil {
		klog.Errorf("Error writing to file %s: %v", args.Path, err)
	}
	return nil
}
