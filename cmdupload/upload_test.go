package cmdupload

import (
	"cmp"
	"context"
	"immich-go/assets"
	"immich-go/helpers/gen"
	"immich-go/immich"
	"immich-go/immich/logger"
	"reflect"
	"slices"
	"testing"

	"github.com/kr/pretty"
)

type stubIC struct {
}

func (c *stubIC) GetAllAssetsWithFilter(context.Context, *immich.GetAssetOptions, func(*immich.Asset)) error {
	return nil
}
func (c *stubIC) AssetUpload(context.Context, *assets.LocalAssetFile) (immich.AssetResponse, error) {
	return immich.AssetResponse{}, nil
}
func (c *stubIC) DeleteAssets(context.Context, []string) error {
	return nil
}
func (c *stubIC) GetAllAlbums(context.Context) ([]immich.AlbumSimplified, error) {
	return nil, nil
}
func (c *stubIC) AddAssetToAlbum(context.Context, string, []string) ([]immich.UpdateAlbumResult, error) {
	return nil, nil
}
func (c *stubIC) CreateAlbum(context.Context, string, []string) (immich.AlbumSimplified, error) {
	return immich.AlbumSimplified{}, nil
}

// type mockedBrowser struct {
// 	assets []assets.LocalAssetFile
// }

// func (m *mockedBrowser) Browse(cxt context.Context) chan *assets.LocalAssetFile {
// 	c := make(chan *assets.LocalAssetFile)
// 	go func() {
// 		for _, a := range m.assets {
// 			c <- &a
// 		}
// 		close(c)
// 	}()
// 	return c
// }

type icCatchUploadsAssets struct {
	stubIC

	assets []string
	albums map[string][]string
}

func (c *icCatchUploadsAssets) AssetUpload(ctx context.Context, a *assets.LocalAssetFile) (immich.AssetResponse, error) {
	c.assets = append(c.assets, a.FileName)
	return immich.AssetResponse{
		ID: a.FileName,
	}, nil
}
func (c *icCatchUploadsAssets) AddAssetToAlbum(ctx context.Context, album string, ids []string) ([]immich.UpdateAlbumResult, error) {
	return nil, nil
}
func (c *icCatchUploadsAssets) CreateAlbum(ctx context.Context, album string, ids []string) (immich.AlbumSimplified, error) {
	if c.albums == nil {
		c.albums = map[string][]string{}
	}
	l := c.albums[album]
	c.albums[album] = append(l, ids...)
	return immich.AlbumSimplified{
		ID:        album,
		AlbumName: album,
	}, nil
}

func TestUpload(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedErr    bool
		expectedAssets []string
		expectedAlbums map[string][]string
	}{
		{
			name: "Simple file",
			args: []string{
				"TEST_DATA/folder/low/PXL_20231006_063000139.jpg",
			},
			expectedErr:    false,
			expectedAssets: []string{"PXL_20231006_063000139.jpg"},
			expectedAlbums: map[string][]string{},
		},
		{
			name: "Simple file in an album",
			args: []string{
				"-album=the album",
				"TEST_DATA/folder/low/PXL_20231006_063000139.jpg",
			},
			expectedErr: false,
			expectedAssets: []string{
				"PXL_20231006_063000139.jpg",
			},
			expectedAlbums: map[string][]string{
				"the album": {"PXL_20231006_063000139.jpg"},
			},
		},
		{
			name: "Folders, no album creation",
			args: []string{
				"TEST_DATA/folder/high",
			},
			expectedErr: false,
			expectedAssets: []string{
				"AlbumA/PXL_20231006_063000139.jpg",
				"AlbumA/PXL_20231006_063029647.jpg",
				"AlbumA/PXL_20231006_063108407.jpg",
				"AlbumA/PXL_20231006_063121958.jpg",
				"AlbumA/PXL_20231006_063357420.jpg",
				"AlbumB/PXL_20231006_063528961.jpg",
				"AlbumB/PXL_20231006_063536303.jpg",
				"AlbumB/PXL_20231006_063851485.jpg",
			},
			expectedAlbums: map[string][]string{},
		},
		{
			name: "Folders, in given album",
			args: []string{
				"-album=the album",
				"TEST_DATA/folder/high",
			},
			expectedErr: false,
			expectedAssets: []string{
				"AlbumA/PXL_20231006_063000139.jpg",
				"AlbumA/PXL_20231006_063029647.jpg",
				"AlbumA/PXL_20231006_063108407.jpg",
				"AlbumA/PXL_20231006_063121958.jpg",
				"AlbumA/PXL_20231006_063357420.jpg",
				"AlbumB/PXL_20231006_063528961.jpg",
				"AlbumB/PXL_20231006_063536303.jpg",
				"AlbumB/PXL_20231006_063851485.jpg",
			},
			expectedAlbums: map[string][]string{
				"the album": {
					"AlbumA/PXL_20231006_063000139.jpg",
					"AlbumA/PXL_20231006_063029647.jpg",
					"AlbumA/PXL_20231006_063108407.jpg",
					"AlbumA/PXL_20231006_063121958.jpg",
					"AlbumA/PXL_20231006_063357420.jpg",
					"AlbumB/PXL_20231006_063528961.jpg",
					"AlbumB/PXL_20231006_063536303.jpg",
					"AlbumB/PXL_20231006_063851485.jpg",
				},
			},
		},
		{
			name: "Folders, album after folder",
			args: []string{
				"-create-album-folder=TRUE",
				"TEST_DATA/folder/high",
			},
			expectedErr: false,
			expectedAssets: []string{
				"AlbumA/PXL_20231006_063000139.jpg",
				"AlbumA/PXL_20231006_063029647.jpg",
				"AlbumA/PXL_20231006_063108407.jpg",
				"AlbumA/PXL_20231006_063121958.jpg",
				"AlbumA/PXL_20231006_063357420.jpg",
				"AlbumB/PXL_20231006_063528961.jpg",
				"AlbumB/PXL_20231006_063536303.jpg",
				"AlbumB/PXL_20231006_063851485.jpg",
			},
			expectedAlbums: map[string][]string{
				"AlbumA": {
					"AlbumA/PXL_20231006_063000139.jpg",
					"AlbumA/PXL_20231006_063029647.jpg",
					"AlbumA/PXL_20231006_063108407.jpg",
					"AlbumA/PXL_20231006_063121958.jpg",
					"AlbumA/PXL_20231006_063357420.jpg",
				},
				"AlbumB": {
					"AlbumB/PXL_20231006_063528961.jpg",
					"AlbumB/PXL_20231006_063536303.jpg",
					"AlbumB/PXL_20231006_063851485.jpg",
				},
			},
		},
		{
			name: "google photos, default options",
			args: []string{
				"-google-photos",
				"TEST_DATA/Takeout1",
			},
			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063000139.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063029647.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063108407.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063121958.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063357420.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063536303.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063851485.jpg",
				"Google Photos/Album test 6-10-23/PXL_20231006_063909898.LS.mp4",
			},
			expectedAlbums: map[string][]string{
				"Album test 6/10/23": {
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063000139.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063029647.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063108407.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063121958.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063357420.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063536303.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063851485.jpg",
					"Google Photos/Album test 6-10-23/PXL_20231006_063909898.LS.mp4",
				},
			},
		},
		{
			name: "google photos, album name from folder",
			args: []string{
				"-google-photos",
				"--use-album-folder-as-name=TRUE",
				"TEST_DATA/Takeout1",
			},
			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063000139.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063029647.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063108407.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063121958.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063357420.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063536303.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063851485.jpg",
				"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063909898.LS.mp4",
			},
			expectedAlbums: map[string][]string{
				"Album test 6-10-23": {
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063000139.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063029647.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063108407.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063121958.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063357420.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063536303.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063851485.jpg",
					"Google\u00a0Photos/Album test 6-10-23/PXL_20231006_063909898.LS.mp4",
				},
			},
		},
		{
			name: "google photo, ignore untitled, discard partner",
			args: []string{
				"-google-photos",
				"--keep-partner=FALSE",
				"TEST_DATA/Takeout2",
			},
			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063528961.jpg",
				"Google\u00a0Photos/Sans titre(9)/PXL_20231006_063108407.jpg",
			},
			expectedAlbums: map[string][]string{},
		},
		{
			name: "google photo, ignore untitled, keep partner",
			args: []string{
				"-google-photos",
				"TEST_DATA/Takeout2",
			},

			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063528961.jpg",
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063000139.jpg",
				"Google\u00a0Photos/Sans titre(9)/PXL_20231006_063108407.jpg",
			},
			expectedAlbums: map[string][]string{},
		},
		{
			name: "google photo, ignore untitled, keep partner, partner album",
			args: []string{
				"-google-photos",
				"-partner-album=partner",
				"TEST_DATA/Takeout2",
			},
			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063528961.jpg",
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063000139.jpg",
				"Google\u00a0Photos/Sans titre(9)/PXL_20231006_063108407.jpg",
			},
			expectedAlbums: map[string][]string{
				"partner": {
					"Google\u00a0Photos/Photos from 2023/PXL_20231006_063000139.jpg",
				},
			},
		},
		{
			name: "google photo, keep untitled",
			args: []string{
				"-google-photos",
				"-keep-untitled-albums=TRUE",
				"-partner-album=partner",
				"TEST_DATA/Takeout2",
			},
			expectedErr: false,
			expectedAssets: []string{
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063528961.jpg",
				"Google\u00a0Photos/Photos from 2023/PXL_20231006_063000139.jpg",
				"Google\u00a0Photos/Sans titre(9)/PXL_20231006_063108407.jpg",
			},
			expectedAlbums: map[string][]string{
				"partner": {
					"Google\u00a0Photos/Photos from 2023/PXL_20231006_063000139.jpg",
				},
				"Sans titre(9)": {
					"Google\u00a0Photos/Sans titre(9)/PXL_20231006_063108407.jpg",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ic := &icCatchUploadsAssets{
				albums: map[string][]string{},
			}
			log := logger.NewLogger(logger.OK, false, false).Writer(nil)
			ctx := context.Background()

			app, err := NewUpCmd(ctx, ic, log, tc.args)
			if err != nil {
				t.Errorf("can't instantiate the UploadCmd: %s", err)
				return
			}

			err = app.Run(ctx)
			if (tc.expectedErr && err == nil) || (!tc.expectedErr && err != nil) {
				t.Errorf("unexpected error condition: %v,%s", tc.expectedErr, err)
				return
			}

			if !cmpSlices(tc.expectedAssets, ic.assets) {
				t.Errorf("expected upload differs ")
				pretty.Ldiff(t, tc.expectedAssets, ic.assets)
			}
			if !cmpAlbums(tc.expectedAlbums, ic.albums) {
				t.Errorf("expected albums differs ")
				pretty.Ldiff(t, tc.expectedAlbums, ic.albums)
			}
		})
	}
}

func cmpAlbums(a, b map[string][]string) bool {
	ka := gen.MapKeys(a)
	kb := gen.MapKeys(b)
	if !cmpSlices(ka, kb) {
		return false
	}
	r := true
	for _, k := range ka {
		r = r && cmpSlices(a[k], b[k])
		if !r {
			return r
		}
	}
	return r
}

func cmpSlices[T cmp.Ordered](a, b []T) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	slices.Sort(a)
	slices.Sort(b)
	return reflect.DeepEqual(a, b)
}