package clouddrive

import (
	"context"
	"fmt"
	"testing"
)

func TestStorageQuota(t *testing.T) {
	cd, err := NewCloudDriveService("/home/tsl4/gocode/src/github.com/rclone/rclone/key.json", context.Background(), new(Options))
	if err != nil {
		t.Errorf("Error: fetching the storage info: %s", err)
	}
	for _, account := range *cd.StorageServiceAccount {
		service, err := account.getDriveService(context.Background(), new(Options))
		if err != nil {
			t.Errorf("Error: fetching the storage info: %s", err)
		}
		abt, err := service.About.Get().Fields("storageQuota").Do()
		if err != nil {
			t.Errorf("Error: fetching the storage info")
		}
		t.Log("======>" + fmt.Sprintf("Key = %s, %dKB", account.Name, uint64(abt.StorageQuota.Usage)/(1024)))
	}
}
