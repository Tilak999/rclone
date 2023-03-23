package clouddrive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/lib/env"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type MasterKey struct {
	IndexStoreKey   string                     `json:"indexStoreKey"`
	ServiceAccounts map[string]json.RawMessage `json:"serviceAccounts"`
}

type ServiceAccount struct {
	ServiceAccountJson []byte
	service            *drive.Service
	client             *http.Client
	//usedStorage        int64
	//freeStorage        int64
	Name        string `json:"key"`
	ClientEmail string `json:"client_email"`
}

type CloudDriveService struct {
	IndexServiceAccount   *ServiceAccount
	StorageServiceAccount *[]ServiceAccount
	opts                  *Options
	serviceAccountMap     map[string]*ServiceAccount
}

var lastServiceAccount *ServiceAccount

func NewCloudDriveService(keyFile string, ctx context.Context, opt *Options) (*CloudDriveService, error) {
	if keyFile == "" {
		return nil, fmt.Errorf("invalid master key file path: %s", keyFile)
	}

	cloudDriveService := new(CloudDriveService)
	loadedCreds, err := os.ReadFile(env.ShellExpand(keyFile))
	if err != nil {
		return nil, fmt.Errorf("error opening master key file: %w", err)
	}

	masterKey := MasterKey{}
	err = json.Unmarshal(loadedCreds, &masterKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing master key file: %w", err)
	}

	cloudDriveService.IndexServiceAccount = &ServiceAccount{
		ServiceAccountJson: []byte(masterKey.ServiceAccounts[masterKey.IndexStoreKey]),
	}
	err = json.Unmarshal(cloudDriveService.IndexServiceAccount.ServiceAccountJson, cloudDriveService.IndexServiceAccount)
	if err != nil {
		return nil, fmt.Errorf("error parsing index service account credentials: %w", err)
	}

	serviceAccountJson := make([]ServiceAccount, 0, len(masterKey.ServiceAccounts)-1)
	for k, v := range masterKey.ServiceAccounts {
		if k != cloudDriveService.IndexServiceAccount.Name {
			temp := new(ServiceAccount)
			json.Unmarshal(v, &temp)
			temp.Name = k
			temp.ServiceAccountJson = v
			serviceAccountJson = append(serviceAccountJson, *temp)
		}
	}
	cloudDriveService.StorageServiceAccount = &serviceAccountJson
	cloudDriveService.opts = opt
	cloudDriveService.IndexServiceAccount.getDriveService(ctx, opt)
	cloudDriveService.serviceAccountMap = make(map[string]*ServiceAccount)
	return cloudDriveService, nil
}

func (s *ServiceAccount) getHttpClientWith(ctx context.Context, opt *Options) (*http.Client, error) {
	if s.client != nil {
		return s.client, nil
	}
	return getServiceAccountClient(context.Background(), new(Options), s.ServiceAccountJson)
}

func (s *ServiceAccount) getDriveService(ctx context.Context, opt *Options) (*drive.Service, error) {
	if s.service != nil {
		return s.service, nil
	}
	client, err := s.getHttpClientWith(ctx, opt)
	if err != nil {
		return nil, err
	}
	return drive.NewService(context.Background(), option.WithHTTPClient(client))
}

func (c *CloudDriveService) getNextServiceAccount(ctx context.Context, size int64) (*ServiceAccount, error) {
	if lastServiceAccount != nil {
		return lastServiceAccount, nil
	}
	for _, storageAccount := range *c.StorageServiceAccount {
		service, err := storageAccount.getDriveService(ctx, c.opts)
		if err != nil {
			return nil, fmt.Errorf("Can't get drive service for"+storageAccount.Name+": %w", err)
		}
		abt, err := service.About.Get().Fields("storageQuota").Do()
		if err != nil {
			return nil, fmt.Errorf("error fetching storage info: %s", err)
		}
		if size < (abt.StorageQuota.Limit - abt.StorageQuota.Usage) {
			lastServiceAccount = &storageAccount
		}
	}
	if lastServiceAccount == nil {
		return nil, fmt.Errorf("storage full no service account file is able to fulfil the request")
	}
	return lastServiceAccount, nil
}

func (c *CloudDriveService) getServiceAccountByName(name string) *ServiceAccount {
	account := c.serviceAccountMap[name]
	if account != nil {
		return account
	}
	for _, acc := range *c.StorageServiceAccount {
		if acc.Name == name {
			c.serviceAccountMap[acc.Name] = &acc
			return &acc
		}
	}
	return nil
}

func (c *CloudDriveService) deleteFile(ctx context.Context, file *drive.File) error {
	fs.Debugf("Deleting file:"+file.Name+", from account", file.Description)
	if file.Description == "" || file.MimeType == driveFolderType {
		svc, err := c.IndexServiceAccount.getDriveService(ctx, c.opts)
		if err == nil {
			err = svc.Files.Delete(file.Id).Fields("").SupportsAllDrives(true).Context(ctx).Do()
		}
		return err
	} else {
		serviceAccount := c.getServiceAccountByName(file.Description)
		svc, err := serviceAccount.getDriveService(ctx, c.opts)
		if err == nil {
			err = svc.Files.Delete(file.Id).Fields("").SupportsAllDrives(true).Context(ctx).Do()
		}
		return err
	}
}

func (c *CloudDriveService) delete(ctx context.Context, f *Fs, id string) (bool, error) {
	fs.Debugf("[cloud drive] fetching file details, id", id)

	var err error
	var file *drive.File

	file, err = f.svc.Files.Get(id).Fields(partialFields).Do()
	if err != nil {
		return false, err
	}

	if file.MimeType == driveFolderType || file.Description == "" {
		_, err = f.list(ctx, []string{file.Id}, file.Name, false, false, false, true, func(child *drive.File) bool {
			c.delete(ctx, f, child.Id)
			return false
		})
		if err != nil {
			return false, fmt.Errorf("[cloud drive] failed to list all the files and dir for deletion: %w", err)
		}

	} else {
		fs.Debugf("[cloud drive] deleting object name", file.Name)
		actualFile := &drive.File{}
		err = json.Unmarshal([]byte(file.Description), actualFile)
		if err != nil {
			return false, fmt.Errorf("[cloud drive] failed to parse json: %w", err)
		}
		err = c.deleteFile(ctx, actualFile)
		if err != nil {
			return false, fmt.Errorf("[cloud drive] failed to delete file:"+actualFile.Name+": %w", err)
		}
	}
	fs.Debugf("[cloud drive] deleting shortcut or dir", file.Name)
	f.svc.Files.Delete(file.Id).SupportsAllDrives(true).Context(ctx).Do()
	return true, nil
}
