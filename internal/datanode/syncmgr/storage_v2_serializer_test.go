// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncmgr

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/msgpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/schemapb"
	milvus_storage "github.com/milvus-io/milvus-storage/go/storage"
	"github.com/milvus-io/milvus-storage/go/storage/options"
	"github.com/milvus-io/milvus-storage/go/storage/schema"
	"github.com/milvus-io/milvus/internal/datanode/metacache"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/storage"
	"github.com/milvus-io/milvus/pkg/common"
	"github.com/milvus-io/milvus/pkg/util/paramtable"
	"github.com/milvus-io/milvus/pkg/util/tsoutil"
)

type StorageV2SerializerSuite struct {
	suite.Suite

	collectionID int64
	partitionID  int64
	segmentID    int64
	channelName  string

	schema         *schemapb.CollectionSchema
	storageCache   *metacache.StorageV2Cache
	mockCache      *metacache.MockMetaCache
	mockMetaWriter *MockMetaWriter

	serializer *storageV2Serializer
}

func (s *StorageV2SerializerSuite) SetupSuite() {
	paramtable.Get().Init(paramtable.NewBaseTable())

	s.collectionID = rand.Int63n(100) + 1000
	s.partitionID = rand.Int63n(100) + 2000
	s.segmentID = rand.Int63n(1000) + 10000
	s.channelName = fmt.Sprintf("by-dev-rootcoord-dml0_%d_v1", s.collectionID)
	s.schema = &schemapb.CollectionSchema{
		Name: "sync_task_test_col",
		Fields: []*schemapb.FieldSchema{
			{FieldID: common.RowIDField, DataType: schemapb.DataType_Int64, Name: common.RowIDFieldName},
			{FieldID: common.TimeStampField, DataType: schemapb.DataType_Int64, Name: common.TimeStampFieldName},
			{
				FieldID:      100,
				Name:         "pk",
				DataType:     schemapb.DataType_Int64,
				IsPrimaryKey: true,
			},
			{
				FieldID:  101,
				Name:     "vector",
				DataType: schemapb.DataType_FloatVector,
				TypeParams: []*commonpb.KeyValuePair{
					{Key: common.DimKey, Value: "128"},
				},
			},
		},
	}

	s.mockCache = metacache.NewMockMetaCache(s.T())
	s.mockMetaWriter = NewMockMetaWriter(s.T())
}

func (s *StorageV2SerializerSuite) SetupTest() {
	storageCache, err := metacache.NewStorageV2Cache(s.schema)
	s.Require().NoError(err)
	s.storageCache = storageCache

	s.mockCache.EXPECT().Collection().Return(s.collectionID)
	s.mockCache.EXPECT().Schema().Return(s.schema)

	s.serializer, err = NewStorageV2Serializer(storageCache, s.mockCache, s.mockMetaWriter)
	s.Require().NoError(err)
}

func (s *StorageV2SerializerSuite) getSpace() *milvus_storage.Space {
	tmpDir := s.T().TempDir()
	space, err := milvus_storage.Open(fmt.Sprintf("file:///%s", tmpDir), options.NewSpaceOptionBuilder().
		SetSchema(schema.NewSchema(s.storageCache.ArrowSchema(), &schema.SchemaOptions{
			PrimaryColumn: "pk", VectorColumn: "vector", VersionColumn: common.TimeStampFieldName,
		})).Build())
	s.Require().NoError(err)
	return space
}

func (s *StorageV2SerializerSuite) getBasicPack() *SyncPack {
	pack := &SyncPack{}

	pack.WithCollectionID(s.collectionID).
		WithPartitionID(s.partitionID).
		WithSegmentID(s.segmentID).
		WithChannelName(s.channelName).
		WithCheckpoint(&msgpb.MsgPosition{
			Timestamp:   1000,
			ChannelName: s.channelName,
		})

	return pack
}

func (s *StorageV2SerializerSuite) getEmptyInsertBuffer() *storage.InsertData {
	buf, err := storage.NewInsertData(s.schema)
	s.Require().NoError(err)

	return buf
}

func (s *StorageV2SerializerSuite) getInsertBuffer() *storage.InsertData {
	buf := s.getEmptyInsertBuffer()

	// generate data
	for i := 0; i < 10; i++ {
		data := make(map[storage.FieldID]any)
		data[common.RowIDField] = int64(i + 1)
		data[common.TimeStampField] = int64(i + 1)
		data[100] = int64(i + 1)
		vector := lo.RepeatBy(128, func(_ int) float32 {
			return rand.Float32()
		})
		data[101] = vector
		err := buf.Append(data)
		s.Require().NoError(err)
	}
	return buf
}

func (s *StorageV2SerializerSuite) getDeleteBuffer() *storage.DeleteData {
	buf := &storage.DeleteData{}
	for i := 0; i < 10; i++ {
		pk := storage.NewInt64PrimaryKey(int64(i + 1))
		ts := tsoutil.ComposeTSByTime(time.Now(), 0)
		buf.Append(pk, ts)
	}
	return buf
}

func (s *StorageV2SerializerSuite) getDeleteBufferZeroTs() *storage.DeleteData {
	buf := &storage.DeleteData{}
	for i := 0; i < 10; i++ {
		pk := storage.NewInt64PrimaryKey(int64(i + 1))
		buf.Append(pk, 0)
	}
	return buf
}

func (s *StorageV2SerializerSuite) getBfs() *metacache.BloomFilterSet {
	bfs := metacache.NewBloomFilterSet()
	fd, err := storage.NewFieldData(schemapb.DataType_Int64, &schemapb.FieldSchema{
		FieldID:      101,
		Name:         "ID",
		IsPrimaryKey: true,
		DataType:     schemapb.DataType_Int64,
	})
	s.Require().NoError(err)

	ids := []int64{1, 2, 3, 4, 5, 6, 7}
	for _, id := range ids {
		err = fd.AppendRow(id)
		s.Require().NoError(err)
	}

	bfs.UpdatePKRange(fd)
	return bfs
}

func (s *StorageV2SerializerSuite) TestSerializeInsert() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.storageCache.SetSpace(s.segmentID, s.getSpace())

	s.Run("no_data", func() {
		pack := s.getBasicPack()
		pack.WithTimeRange(50, 100)
		pack.WithDrop()

		task, err := s.serializer.EncodeBuffer(ctx, pack)
		s.NoError(err)
		taskV1, ok := task.(*SyncTaskV2)
		s.Require().True(ok)
		s.Equal(s.collectionID, taskV1.collectionID)
		s.Equal(s.partitionID, taskV1.partitionID)
		s.Equal(s.channelName, taskV1.channelName)
		s.Equal(&msgpb.MsgPosition{
			Timestamp:   1000,
			ChannelName: s.channelName,
		}, taskV1.checkpoint)
		s.EqualValues(50, taskV1.tsFrom)
		s.EqualValues(100, taskV1.tsTo)
		s.True(taskV1.isDrop)
	})

	s.Run("empty_insert_data", func() {
		pack := s.getBasicPack()
		pack.WithTimeRange(50, 100)
		pack.WithInsertData(s.getEmptyInsertBuffer()).WithBatchSize(0)

		_, err := s.serializer.EncodeBuffer(ctx, pack)
		s.Error(err)
	})

	s.Run("with_normal_data", func() {
		pack := s.getBasicPack()
		pack.WithTimeRange(50, 100)
		pack.WithInsertData(s.getInsertBuffer()).WithBatchSize(10)

		s.mockCache.EXPECT().UpdateSegments(mock.Anything, mock.Anything).Return().Once()

		task, err := s.serializer.EncodeBuffer(ctx, pack)
		s.NoError(err)

		taskV2, ok := task.(*SyncTaskV2)
		s.Require().True(ok)
		s.Equal(s.collectionID, taskV2.collectionID)
		s.Equal(s.partitionID, taskV2.partitionID)
		s.Equal(s.channelName, taskV2.channelName)
		s.Equal(&msgpb.MsgPosition{
			Timestamp:   1000,
			ChannelName: s.channelName,
		}, taskV2.checkpoint)
		s.EqualValues(50, taskV2.tsFrom)
		s.EqualValues(100, taskV2.tsTo)
		s.NotNil(taskV2.reader)
		s.NotNil(taskV2.batchStatsBlob)
	})

	s.Run("with_flush_segment_not_found", func() {
		pack := s.getBasicPack()
		pack.WithFlush()

		s.mockCache.EXPECT().GetSegmentByID(s.segmentID).Return(nil, false).Once()
		_, err := s.serializer.EncodeBuffer(ctx, pack)
		s.Error(err)
	})

	s.Run("with_flush", func() {
		pack := s.getBasicPack()
		pack.WithTimeRange(50, 100)
		pack.WithInsertData(s.getInsertBuffer()).WithBatchSize(10)
		pack.WithFlush()

		bfs := s.getBfs()
		segInfo := metacache.NewSegmentInfo(&datapb.SegmentInfo{}, bfs)
		metacache.UpdateNumOfRows(1000)(segInfo)
		metacache.CompactTo(metacache.NullSegment)(segInfo)
		s.mockCache.EXPECT().UpdateSegments(mock.Anything, mock.Anything).Run(func(action metacache.SegmentAction, filters ...metacache.SegmentFilter) {
			action(segInfo)
		}).Return().Once()
		s.mockCache.EXPECT().GetSegmentByID(s.segmentID).Return(segInfo, true).Once()

		task, err := s.serializer.EncodeBuffer(ctx, pack)
		s.NoError(err)

		taskV2, ok := task.(*SyncTaskV2)
		s.Require().True(ok)
		s.Equal(s.collectionID, taskV2.collectionID)
		s.Equal(s.partitionID, taskV2.partitionID)
		s.Equal(s.channelName, taskV2.channelName)
		s.Equal(&msgpb.MsgPosition{
			Timestamp:   1000,
			ChannelName: s.channelName,
		}, taskV2.checkpoint)
		s.EqualValues(50, taskV2.tsFrom)
		s.EqualValues(100, taskV2.tsTo)
		s.NotNil(taskV2.mergedStatsBlob)
	})
}

func (s *StorageV2SerializerSuite) TestSerializeDelete() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Run("serialize_failed", func() {
		pkField := s.serializer.pkField
		s.serializer.pkField = &schemapb.FieldSchema{}
		defer func() {
			s.serializer.pkField = pkField
		}()
		pack := s.getBasicPack()
		pack.WithDeleteData(s.getDeleteBufferZeroTs())
		pack.WithTimeRange(50, 100)

		_, err := s.serializer.EncodeBuffer(ctx, pack)
		s.Error(err)
	})

	s.Run("serialize_failed_bad_pk", func() {
		pkField := s.serializer.pkField
		s.serializer.pkField = &schemapb.FieldSchema{
			DataType: schemapb.DataType_Array,
		}
		defer func() {
			s.serializer.pkField = pkField
		}()
		pack := s.getBasicPack()
		pack.WithDeleteData(s.getDeleteBufferZeroTs())
		pack.WithTimeRange(50, 100)

		_, err := s.serializer.EncodeBuffer(ctx, pack)
		s.Error(err)
	})

	s.Run("serialize_normal", func() {
		pack := s.getBasicPack()
		pack.WithDeleteData(s.getDeleteBuffer())
		pack.WithTimeRange(50, 100)

		task, err := s.serializer.EncodeBuffer(ctx, pack)
		s.NoError(err)

		taskV2, ok := task.(*SyncTaskV2)
		s.Require().True(ok)
		s.Equal(s.collectionID, taskV2.collectionID)
		s.Equal(s.partitionID, taskV2.partitionID)
		s.Equal(s.channelName, taskV2.channelName)
		s.Equal(&msgpb.MsgPosition{
			Timestamp:   1000,
			ChannelName: s.channelName,
		}, taskV2.checkpoint)
		s.EqualValues(50, taskV2.tsFrom)
		s.EqualValues(100, taskV2.tsTo)
		s.NotNil(taskV2.deleteReader)
	})
}

func (s *StorageV2SerializerSuite) TestBadSchema() {
	mockCache := metacache.NewMockMetaCache(s.T())
	mockCache.EXPECT().Collection().Return(s.collectionID).Once()
	mockCache.EXPECT().Schema().Return(&schemapb.CollectionSchema{}).Once()
	_, err := NewStorageV2Serializer(s.storageCache, mockCache, s.mockMetaWriter)
	s.Error(err)
}

func TestStorageV2Serializer(t *testing.T) {
	suite.Run(t, new(StorageV2SerializerSuite))
}
