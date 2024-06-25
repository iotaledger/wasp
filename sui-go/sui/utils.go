package sui

import (
	"fmt"

	"github.com/iotaledger/wasp/sui-go/models"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

// requires `ShowObjectChanges: true`
func GetCreatedObjectRefAndType(
	txRes *models.SuiTransactionBlockResponse,
	moduleName string,
	objectName string,
) (*sui_types.ObjectRef, string, error) {
	if txRes.ObjectChanges == nil {
		return nil, "", fmt.Errorf("no ObjectChanges")
	}
	for _, change := range txRes.ObjectChanges {
		fmt.Println("change.Data.String(): ", change.Data.String())
		if change.Data.Created != nil {
			// some possible examples
			// * 0x2::coin::TreasuryCap<0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::testcoin::TESTCOIN>
			// * 0x14c12b454ac6996024342312769e00bb98c70ad2f3546a40f62516c83aa0f0d4::anchor::Anchor
			resource, err := models.NewResourceType(change.Data.Created.ObjectType)
			if err != nil {
				return nil, "", fmt.Errorf("invalid resource string")
			}
			if resource.Module == moduleName && resource.ObjectName == objectName {
				ref := &sui_types.ObjectRef{
					ObjectID: &change.Data.Created.ObjectID,
					Version:  change.Data.Created.Version.Uint64(),
					Digest:   &change.Data.Created.Digest,
				}
				return ref, change.Data.Created.ObjectType, nil
			}
			for ; resource != nil; resource = resource.SubType {
				if resource.Module == moduleName && resource.ObjectName == objectName {
					ref := &sui_types.ObjectRef{
						ObjectID: &change.Data.Created.ObjectID,
						Version:  change.Data.Created.Version.Uint64(),
						Digest:   &change.Data.Created.Digest,
					}
					return ref, change.Data.Created.ObjectType, nil
				}
			}
		}
	}
	return nil, "", fmt.Errorf("not found")
}
