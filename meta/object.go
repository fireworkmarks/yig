package meta

import (
	. "database/sql/driver"

	. "github.com/journeymidnight/yig/context"
	. "github.com/journeymidnight/yig/error"
	"github.com/journeymidnight/yig/helper"
	. "github.com/journeymidnight/yig/meta/types"
	"github.com/journeymidnight/yig/redis"
)

func (m *Meta) GetObject(bucketName string, objectName string, willNeed bool) (object *Object, err error) {
	//getObject := func() (o interface{}, err error) {
	//	helper.Logger.Info("GetObject CacheMiss. bucket:", bucketName,
	//		"object:", objectName)
	//	object, err := m.Client.GetObject(bucketName, objectName, "")
	//	if err != nil {
	//		return
	//	}
	//	helper.Logger.Info("GetObject object.Name:", object.Name)
	//	if object.Name != objectName {
	//		err = ErrNoSuchKey
	//		return
	//	}
	//	return object, nil
	//}
	//unmarshaller := func(in []byte) (interface{}, error) {
	//	var object Object
	//	err := helper.MsgPackUnMarshal(in, &object)
	//	return &object, err
	//}
	//
	//o, err := m.Cache.Get(redis.ObjectTable, bucketName+":"+objectName+":",
	//	getObject, unmarshaller, willNeed)
	//if err != nil {
	//	return
	//}
	//object, ok := o.(*Object)
	//if !ok {
	//	err = ErrInternalError
	//	return
	//}
	object = new(Object)
	return object, nil
}

func (m *Meta) GetObjectVersion(bucketName, objectName, version string, willNeed bool) (object *Object, err error) {
	getObjectVersion := func() (o interface{}, err error) {
		object, err := m.Client.GetObject(bucketName, objectName, version)
		if err != nil {
			return
		}
		if object.Name != objectName {
			err = ErrNoSuchKey
			return
		}
		return object, nil
	}
	unmarshaller := func(in []byte) (interface{}, error) {
		var object Object
		err := helper.MsgPackUnMarshal(in, &object)
		return &object, err
	}
	o, err := m.Cache.Get(redis.ObjectTable, bucketName+":"+objectName+":"+version,
		getObjectVersion, unmarshaller, willNeed)
	if err != nil {
		return
	}
	object, ok := o.(*Object)
	if !ok {
		err = ErrInternalError
		return
	}
	return object, nil
}

func (m *Meta) PutObject(reqCtx RequestContext, object *Object, multipart *Multipart, updateUsage bool) error {
	if reqCtx.BucketInfo == nil {
		return ErrNoSuchBucket
	}
	//if reqCtx.BucketInfo.Versioning == VersionDisabled {
	//	object.VersionId = NullVersion
	//} else {
	//	return ErrNotImplemented
	//	// TODO: object.VersionId = strconv.FormatUint(math.MaxUint64-uint64(object.LastModifiedTime.UnixNano()), 10)
	//}

	needUpdate := (reqCtx.ObjectInfo != nil)
	if multipart == nil && object.Parts == nil {
		if needUpdate {
			return nil
			//return m.Client.UpdateObjectWithoutMultiPart(object)
		} else {
			return nil
			//return m.Client.PutObjectWithoutMultiPart(object)
		}
	}

	if needUpdate {
		return nil
		//return m.Client.UpdateObject(object, multipart, updateUsage)
	} else {
		return nil
		//return m.Client.PutObject(object, multipart, updateUsage)
	}

	return nil
}

func (m *Meta) UpdateGlacierObject(targetObject, sourceObject *Object, isFreezer bool) (err error) {
	var tx Tx
	tx, err = m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = m.Client.CommitTrans(tx)
		}
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	if isFreezer {
		err = m.Client.UpdateFreezerObject(targetObject, tx)
		if err != nil {
			return err
		}

		err = m.Client.DeleteFreezer(sourceObject.BucketName, sourceObject.Name, tx)
		if err != nil {
			return err
		}
	} else {
		err = m.Client.PutObject(targetObject, nil, true)
		if err != nil {
			return err
		}
	}

	err = m.Client.PutObjectToGarbageCollection(sourceObject, tx)
	if err != nil {
		return err
	}

	return err
}

func (m *Meta) UpdateObjectAcl(object *Object) error {
	err := m.Client.UpdateObjectAcl(object)
	return err
}

func (m *Meta) UpdateObjectAttrs(object *Object) error {
	err := m.Client.UpdateObjectAttrs(object)
	return err
}

func (m *Meta) RenameObject(object *Object, sourceObject string) error {
	err := m.Client.RenameObject(object, sourceObject)
	return err
}

func (m *Meta) ReplaceObjectMetas(object *Object) error {
	err := m.Client.ReplaceObjectMetas(object, nil)
	return err
}

func (m *Meta) DeleteOldObject(object *Object) (err error) {
	tx, err := m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = m.Client.CommitTrans(tx)
		}
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	return m.Client.UpdateUsage(object.BucketName, -object.Size, tx)
}

func (m *Meta) DeleteObject(object *Object) (err error) {
	tx, err := m.Client.NewTrans()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = m.Client.CommitTrans(tx)
		}
		if err != nil {
			m.Client.AbortTrans(tx)
		}
	}()

	err = m.Client.DeleteObject(object, tx)
	if err != nil {
		return err
	}

	err = m.Client.PutObjectToGarbageCollection(object, tx)
	if err != nil {
		return err
	}

	return m.Client.UpdateUsage(object.BucketName, -object.Size, tx)
}

func (m *Meta) AppendObject(object *Object, isExist bool) error {
	if !isExist {
		return m.Client.PutObject(object, nil, true)
	} else {
		return m.Client.UpdateAppendObject(object)
	}
}
