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

package querynode

/*

#cgo CFLAGS: -I${SRCDIR}/../core/output/include

#cgo darwin LDFLAGS: -L${SRCDIR}/../core/output/lib -lmilvus_segcore -Wl,-rpath,"${SRCDIR}/../core/output/lib"
#cgo linux LDFLAGS: -L${SRCDIR}/../core/output/lib -lmilvus_segcore -Wl,-rpath=${SRCDIR}/../core/output/lib

#include "segcore/collection_c.h"
#include "segcore/segment_c.h"

*/
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/common"
	etcdkv "github.com/milvus-io/milvus/internal/kv/etcd"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/metrics"
	"github.com/milvus-io/milvus/internal/proto/datapb"
	"github.com/milvus-io/milvus/internal/proto/internalpb"
	"github.com/milvus-io/milvus/internal/proto/querypb"
	"github.com/milvus-io/milvus/internal/proto/schemapb"
)

// ReplicaInterface specifies all the methods that the Collection object needs to implement in QueryNode.
// In common cases, the system has multiple query nodes. The full data of a collection will be distributed
// across multiple query nodes, and each query node's collectionReplica will maintain its own part.
type ReplicaInterface interface {
	// collection
	// getCollectionIDs returns all collection ids in the collectionReplica
	getCollectionIDs() []UniqueID
	// addCollection creates a new collection and add it to collectionReplica
	addCollection(collectionID UniqueID, schema *schemapb.CollectionSchema) *Collection
	// removeCollection removes the collection from collectionReplica
	removeCollection(collectionID UniqueID) error
	// getCollectionByID gets the collection which id is collectionID
	getCollectionByID(collectionID UniqueID) (*Collection, error)
	// hasCollection checks if collectionReplica has the collection which id is collectionID
	hasCollection(collectionID UniqueID) bool
	// getCollectionNum returns num of collections in collectionReplica
	getCollectionNum() int
	// getPartitionIDs returns partition ids of collection
	getPartitionIDs(collectionID UniqueID) ([]UniqueID, error)
	// getVecFieldIDsByCollectionID returns vector field ids of collection
	getVecFieldIDsByCollectionID(collectionID UniqueID) ([]FieldID, error)
	// getPKFieldIDsByCollectionID returns vector field ids of collection
	getPKFieldIDByCollectionID(collectionID UniqueID) (FieldID, error)
	// getSegmentInfosByColID return segments info by collectionID
	getSegmentInfosByColID(collectionID UniqueID) ([]*querypb.SegmentInfo, error)

	// partition
	// addPartition adds a new partition to collection
	addPartition(collectionID UniqueID, partitionID UniqueID) error
	// removePartition removes the partition from collectionReplica
	removePartition(partitionID UniqueID) error
	// getPartitionByID returns the partition which id is partitionID
	getPartitionByID(partitionID UniqueID) (*Partition, error)
	// hasPartition returns true if collectionReplica has the partition, false otherwise
	hasPartition(partitionID UniqueID) bool
	// getPartitionNum returns num of partitions
	getPartitionNum() int
	// getSegmentIDs returns segment ids
	getSegmentIDs(partitionID UniqueID) ([]UniqueID, error)
	// getSegmentIDsByVChannel returns segment ids which virtual channel is vChannel
	getSegmentIDsByVChannel(partitionID UniqueID, vChannel Channel) ([]UniqueID, error)

	// segment
	// addSegment add a new segment to collectionReplica
	addSegment(segmentID UniqueID, partitionID UniqueID, collectionID UniqueID, vChannelID Channel, segType segmentType, onService bool) error
	// setSegment adds a segment to collectionReplica
	setSegment(segment *Segment) error
	// removeSegment removes a segment from collectionReplica
	removeSegment(segmentID UniqueID) error
	// getSegmentByID returns the segment which id is segmentID
	getSegmentByID(segmentID UniqueID) (*Segment, error)
	// hasSegment returns true if collectionReplica has the segment, false otherwise
	hasSegment(segmentID UniqueID) bool
	// getSegmentNum returns num of segments in collectionReplica
	getSegmentNum() int
	//  getSegmentStatistics returns the statistics of segments in collectionReplica
	getSegmentStatistics() []*internalpb.SegmentStats

	// excluded segments
	//  removeExcludedSegments will remove excludedSegments from collectionReplica
	removeExcludedSegments(collectionID UniqueID)
	// addExcludedSegments will add excludedSegments to collectionReplica
	addExcludedSegments(collectionID UniqueID, segmentInfos []*datapb.SegmentInfo)
	// getExcludedSegments returns excludedSegments of collectionReplica
	getExcludedSegments(collectionID UniqueID) ([]*datapb.SegmentInfo, error)

	// query mu
	// queryLock guards query and delete operations
	queryLock()
	// queryUnlock guards query and delete segment operations
	queryUnlock()
	// queryRLock guards query and delete segment operations
	queryRLock()
	// queryRUnlock guards query and delete segment operations
	queryRUnlock()

	// getSegmentsMemSize get the memory size in bytes of all the Segments
	getSegmentsMemSize() int64
	// freeAll will free all meta info from collectionReplica
	freeAll()
	// printReplica prints the collections, partitions and segments in the collectionReplica
	printReplica()
}

// collectionReplica is the data replication of memory data in query node.
// It implements `ReplicaInterface` interface.
type metaReplica struct {
	mu          sync.RWMutex // guards all
	collections map[UniqueID]*Collection
	partitions  map[UniqueID]*Partition
	segments    map[UniqueID]*Segment

	queryMu          sync.RWMutex
	excludedSegments map[UniqueID][]*datapb.SegmentInfo // map[collectionID]segmentIDs

	etcdKV *etcdkv.EtcdKV
}

// queryLock guards query and delete operations
func (replica *metaReplica) queryLock() {
	replica.queryMu.Lock()
}

// queryUnlock guards query and delete segment operations
func (replica *metaReplica) queryUnlock() {
	replica.queryMu.Unlock()
}

// queryRLock guards query and delete segment operations
func (replica *metaReplica) queryRLock() {
	replica.queryMu.RLock()
}

// queryRUnlock guards query and delete segment operations
func (replica *metaReplica) queryRUnlock() {
	replica.queryMu.RUnlock()
}

// getSegmentsMemSize get the memory size in bytes of all the Segments
func (replica *metaReplica) getSegmentsMemSize() int64 {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	memSize := int64(0)
	for _, segment := range replica.segments {
		memSize += segment.getMemSize()
	}
	return memSize
}

// printReplica prints the collections, partitions and segments in the collectionReplica
func (replica *metaReplica) printReplica() {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	log.Info("collections in collectionReplica", zap.Any("info", replica.collections))
	log.Info("partitions in collectionReplica", zap.Any("info", replica.partitions))
	log.Info("segments in collectionReplica", zap.Any("info", replica.segments))
	log.Info("excludedSegments in collectionReplica", zap.Any("info", replica.excludedSegments))
}

//----------------------------------------------------------------------------------------------------- collection
// getCollectionIDs gets all the collection ids in the collectionReplica
func (replica *metaReplica) getCollectionIDs() []UniqueID {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	collectionIDs := make([]UniqueID, 0)
	for id := range replica.collections {
		collectionIDs = append(collectionIDs, id)
	}
	return collectionIDs
}

// addCollection creates a new collection and add it to collectionReplica
func (replica *metaReplica) addCollection(collectionID UniqueID, schema *schemapb.CollectionSchema) *Collection {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	if col, ok := replica.collections[collectionID]; ok {
		return col
	}

	var newCollection = newCollection(collectionID, schema)
	replica.collections[collectionID] = newCollection
	metrics.QueryNodeNumCollections.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Set(float64(len(replica.collections)))
	return newCollection
}

// removeCollection removes the collection from collectionReplica
func (replica *metaReplica) removeCollection(collectionID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	return replica.removeCollectionPrivate(collectionID)
}

// removeCollectionPrivate is the private function in collectionReplica, to remove collection from collectionReplica
func (replica *metaReplica) removeCollectionPrivate(collectionID UniqueID) error {
	collection, err := replica.getCollectionByIDPrivate(collectionID)
	if err != nil {
		return err
	}

	// block incoming search&query
	collection.Lock()
	defer collection.Unlock()

	// delete partitions
	for _, partitionID := range collection.partitionIDs {
		// ignore error, try to delete
		_ = replica.removePartitionPrivate(partitionID, true)
	}

	deleteCollection(collection)
	delete(replica.collections, collectionID)

	metrics.QueryNodeNumCollections.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Set(float64(len(replica.collections)))
	metrics.QueryNodeNumPartitions.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Sub(float64(len(collection.partitionIDs)))
	return nil
}

// getCollectionByID gets the collection which id is collectionID
func (replica *metaReplica) getCollectionByID(collectionID UniqueID) (*Collection, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.getCollectionByIDPrivate(collectionID)
}

// getCollectionByIDPrivate is the private function in collectionReplica, to get collection from collectionReplica
func (replica *metaReplica) getCollectionByIDPrivate(collectionID UniqueID) (*Collection, error) {
	collection, ok := replica.collections[collectionID]
	if !ok {
		return nil, fmt.Errorf("collection hasn't been loaded or has been released, collection id = %d", collectionID)
	}

	return collection, nil
}

// hasCollection checks if collectionReplica has the collection which id is collectionID
func (replica *metaReplica) hasCollection(collectionID UniqueID) bool {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.hasCollectionPrivate(collectionID)
}

// hasCollectionPrivate is the private function in collectionReplica, to check collection in collectionReplica
func (replica *metaReplica) hasCollectionPrivate(collectionID UniqueID) bool {
	_, ok := replica.collections[collectionID]
	return ok
}

// getCollectionNum returns num of collections in collectionReplica
func (replica *metaReplica) getCollectionNum() int {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return len(replica.collections)
}

// getPartitionIDs returns partition ids of collection
func (replica *metaReplica) getPartitionIDs(collectionID UniqueID) ([]UniqueID, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	collection, err := replica.getCollectionByIDPrivate(collectionID)
	if err != nil {
		return nil, err
	}

	return collection.partitionIDs, nil
}

func (replica *metaReplica) getIndexedFieldIDByCollectionIDPrivate(collectionID UniqueID, segment *Segment) ([]FieldID, error) {
	fields, err := replica.getFieldsByCollectionIDPrivate(collectionID)
	if err != nil {
		return nil, err
	}

	fieldIDS := make([]FieldID, 0)
	for _, field := range fields {
		if segment.hasLoadIndexForIndexedField(field.FieldID) {
			fieldIDS = append(fieldIDS, field.FieldID)
		}
	}
	return fieldIDS, nil
}

func (replica *metaReplica) getVecFieldIDsByCollectionIDPrivate(collectionID UniqueID) ([]FieldID, error) {
	fields, err := replica.getFieldsByCollectionIDPrivate(collectionID)
	if err != nil {
		return nil, err
	}

	vecFields := make([]FieldID, 0)
	for _, field := range fields {
		if field.DataType == schemapb.DataType_BinaryVector || field.DataType == schemapb.DataType_FloatVector {
			vecFields = append(vecFields, field.FieldID)
		}
	}
	return vecFields, nil
}

// getVecFieldIDsByCollectionID returns vector field ids of collection
func (replica *metaReplica) getVecFieldIDsByCollectionID(collectionID UniqueID) ([]FieldID, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	return replica.getVecFieldIDsByCollectionIDPrivate(collectionID)
}

// getPKFieldIDsByCollectionID returns vector field ids of collection
func (replica *metaReplica) getPKFieldIDByCollectionID(collectionID UniqueID) (FieldID, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	fields, err := replica.getFieldsByCollectionIDPrivate(collectionID)
	if err != nil {
		return common.InvalidFieldID, err
	}

	for _, field := range fields {
		if field.IsPrimaryKey {
			return field.FieldID, nil
		}
	}
	return common.InvalidFieldID, nil
}

// getFieldsByCollectionIDPrivate is the private function in collectionReplica, to return vector field ids of collection
func (replica *metaReplica) getFieldsByCollectionIDPrivate(collectionID UniqueID) ([]*schemapb.FieldSchema, error) {
	collection, err := replica.getCollectionByIDPrivate(collectionID)
	if err != nil {
		return nil, err
	}

	if len(collection.Schema().Fields) <= 0 {
		return nil, errors.New("no field in collection %d" + strconv.FormatInt(collectionID, 10))
	}

	return collection.Schema().Fields, nil
}

// getSegmentInfosByColID return segments info by collectionID
func (replica *metaReplica) getSegmentInfosByColID(collectionID UniqueID) ([]*querypb.SegmentInfo, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	segmentInfos := make([]*querypb.SegmentInfo, 0)
	collection, ok := replica.collections[collectionID]
	if !ok {
		// collection not exist, so result segmentInfos is empty
		return segmentInfos, nil
	}

	for _, partitionID := range collection.partitionIDs {
		partition, ok := replica.partitions[partitionID]
		if !ok {
			return nil, fmt.Errorf("the meta of collection %d and partition %d are inconsistent in QueryNode", collectionID, partitionID)
		}
		for _, segmentID := range partition.segmentIDs {
			segment, ok := replica.segments[segmentID]
			if !ok {
				return nil, fmt.Errorf("the meta of partition %d and segment %d are inconsistent in QueryNode", partitionID, segmentID)
			}
			segmentInfo := replica.getSegmentInfo(segment)
			segmentInfos = append(segmentInfos, segmentInfo)
		}
	}

	return segmentInfos, nil
}

//----------------------------------------------------------------------------------------------------- partition
// addPartition adds a new partition to collection
func (replica *metaReplica) addPartition(collectionID UniqueID, partitionID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	return replica.addPartitionPrivate(collectionID, partitionID)
}

// addPartitionPrivate is the private function in collectionReplica, to add a new partition to collection
func (replica *metaReplica) addPartitionPrivate(collectionID UniqueID, partitionID UniqueID) error {
	collection, err := replica.getCollectionByIDPrivate(collectionID)
	if err != nil {
		return err
	}

	if !replica.hasPartitionPrivate(partitionID) {
		collection.addPartitionID(partitionID)
		var newPartition = newPartition(collectionID, partitionID)
		replica.partitions[partitionID] = newPartition
	}

	metrics.QueryNodeNumPartitions.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Set(float64(len(replica.partitions)))
	return nil
}

// removePartition removes the partition from collectionReplica
func (replica *metaReplica) removePartition(partitionID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	return replica.removePartitionPrivate(partitionID, false)
}

// removePartitionPrivate is the private function in collectionReplica, to remove the partition from collectionReplica
// `locked` flag indicates whether corresponding collection lock is accquired before calling this method
func (replica *metaReplica) removePartitionPrivate(partitionID UniqueID, locked bool) error {
	partition, err := replica.getPartitionByIDPrivate(partitionID)
	if err != nil {
		return err
	}

	collection, err := replica.getCollectionByIDPrivate(partition.collectionID)
	if err != nil {
		return err
	}

	if !locked {
		collection.Lock()
		defer collection.Unlock()
	}

	// delete segments
	for _, segmentID := range partition.segmentIDs {
		// try to delete, ignore error
		_ = replica.removeSegmentPrivate(segmentID)
	}

	collection.removePartitionID(partitionID)
	delete(replica.partitions, partitionID)

	metrics.QueryNodeNumPartitions.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Set(float64(len(replica.partitions)))
	return nil
}

// getPartitionByID returns the partition which id is partitionID
func (replica *metaReplica) getPartitionByID(partitionID UniqueID) (*Partition, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.getPartitionByIDPrivate(partitionID)
}

// getPartitionByIDPrivate is the private function in collectionReplica, to get the partition
func (replica *metaReplica) getPartitionByIDPrivate(partitionID UniqueID) (*Partition, error) {
	partition, ok := replica.partitions[partitionID]
	if !ok {
		return nil, fmt.Errorf("partition %d hasn't been loaded or has been released", partitionID)
	}

	return partition, nil
}

// hasPartition returns true if collectionReplica has the partition, false otherwise
func (replica *metaReplica) hasPartition(partitionID UniqueID) bool {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.hasPartitionPrivate(partitionID)
}

// hasPartitionPrivate is the private function in collectionReplica, to check if collectionReplica has the partition
func (replica *metaReplica) hasPartitionPrivate(partitionID UniqueID) bool {
	_, ok := replica.partitions[partitionID]
	return ok
}

// getPartitionNum returns num of partitions
func (replica *metaReplica) getPartitionNum() int {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return len(replica.partitions)
}

// getSegmentIDs returns segment ids
func (replica *metaReplica) getSegmentIDs(partitionID UniqueID) ([]UniqueID, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.getSegmentIDsPrivate(partitionID)
}

// getSegmentIDsByVChannel returns segment ids which virtual channel is vChannel
func (replica *metaReplica) getSegmentIDsByVChannel(partitionID UniqueID, vChannel Channel) ([]UniqueID, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	segmentIDs, err := replica.getSegmentIDsPrivate(partitionID)
	if err != nil {
		return nil, err
	}
	segmentIDsTmp := make([]UniqueID, 0)
	for _, segmentID := range segmentIDs {
		segment, err := replica.getSegmentByIDPrivate(segmentID)
		if err != nil {
			return nil, err
		}
		if segment.vChannelID == vChannel {
			segmentIDsTmp = append(segmentIDsTmp, segment.ID())
		}
	}

	return segmentIDsTmp, nil
}

// getSegmentIDsPrivate is private function in collectionReplica, it returns segment ids
func (replica *metaReplica) getSegmentIDsPrivate(partitionID UniqueID) ([]UniqueID, error) {
	partition, err2 := replica.getPartitionByIDPrivate(partitionID)
	if err2 != nil {
		return nil, err2
	}
	return partition.segmentIDs, nil
}

//----------------------------------------------------------------------------------------------------- segment
// addSegment add a new segment to collectionReplica
func (replica *metaReplica) addSegment(segmentID UniqueID, partitionID UniqueID, collectionID UniqueID, vChannelID Channel, segType segmentType, onService bool) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	collection, err := replica.getCollectionByIDPrivate(collectionID)
	if err != nil {
		return err
	}
	seg, err := newSegment(collection, segmentID, partitionID, collectionID, vChannelID, segType, onService)
	if err != nil {
		return err
	}
	return replica.addSegmentPrivate(segmentID, partitionID, seg)
}

// addSegmentPrivate is private function in collectionReplica, to add a new segment to collectionReplica
func (replica *metaReplica) addSegmentPrivate(segmentID UniqueID, partitionID UniqueID, segment *Segment) error {
	partition, err := replica.getPartitionByIDPrivate(partitionID)
	if err != nil {
		return err
	}

	if replica.hasSegmentPrivate(segmentID) {
		return nil
	}
	partition.addSegmentID(segmentID)
	replica.segments[segmentID] = segment

	metrics.QueryNodeNumSegments.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Inc()
	return nil
}

// setSegment adds a segment to collectionReplica
func (replica *metaReplica) setSegment(segment *Segment) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	_, err := replica.getCollectionByIDPrivate(segment.collectionID)
	if err != nil {
		return err
	}
	return replica.addSegmentPrivate(segment.segmentID, segment.partitionID, segment)
}

// removeSegment removes a segment from collectionReplica
func (replica *metaReplica) removeSegment(segmentID UniqueID) error {
	replica.mu.Lock()
	defer replica.mu.Unlock()
	return replica.removeSegmentPrivate(segmentID)
}

// removeSegmentPrivate is private function in collectionReplica, to remove a segment from collectionReplica
func (replica *metaReplica) removeSegmentPrivate(segmentID UniqueID) error {
	segment, err := replica.getSegmentByIDPrivate(segmentID)
	if err != nil {
		return err
	}

	partition, err2 := replica.getPartitionByIDPrivate(segment.partitionID)
	if err2 != nil {
		return err
	}

	partition.removeSegmentID(segmentID)
	delete(replica.segments, segmentID)
	deleteSegment(segment)

	metrics.QueryNodeNumSegments.WithLabelValues(fmt.Sprint(Params.QueryNodeCfg.GetNodeID())).Dec()
	return nil
}

// getSegmentByID returns the segment which id is segmentID
func (replica *metaReplica) getSegmentByID(segmentID UniqueID) (*Segment, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.getSegmentByIDPrivate(segmentID)
}

// getSegmentByIDPrivate is private function in collectionReplica, it returns the segment which id is segmentID
func (replica *metaReplica) getSegmentByIDPrivate(segmentID UniqueID) (*Segment, error) {
	segment, ok := replica.segments[segmentID]
	if !ok {
		return nil, fmt.Errorf("cannot find segment %d in QueryNode", segmentID)
	}

	return segment, nil
}

// hasSegment returns true if collectionReplica has the segment, false otherwise
func (replica *metaReplica) hasSegment(segmentID UniqueID) bool {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return replica.hasSegmentPrivate(segmentID)
}

// hasSegmentPrivate is private function in collectionReplica, to check if collectionReplica has the segment
func (replica *metaReplica) hasSegmentPrivate(segmentID UniqueID) bool {
	_, ok := replica.segments[segmentID]
	return ok
}

// getSegmentNum returns num of segments in collectionReplica
func (replica *metaReplica) getSegmentNum() int {
	replica.mu.RLock()
	defer replica.mu.RUnlock()
	return len(replica.segments)
}

//  getSegmentStatistics returns the statistics of segments in collectionReplica
func (replica *metaReplica) getSegmentStatistics() []*internalpb.SegmentStats {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	var statisticData = make([]*internalpb.SegmentStats, 0)

	for segmentID, segment := range replica.segments {
		currentMemSize := segment.getMemSize()
		segment.lastMemSize = currentMemSize
		segmentNumOfRows := segment.getRowCount()

		stat := internalpb.SegmentStats{
			SegmentID:        segmentID,
			MemorySize:       currentMemSize,
			NumRows:          segmentNumOfRows,
			RecentlyModified: segment.getRecentlyModified(),
		}

		statisticData = append(statisticData, &stat)
		segment.setRecentlyModified(false)
	}

	return statisticData
}

//  removeExcludedSegments will remove excludedSegments from collectionReplica
func (replica *metaReplica) removeExcludedSegments(collectionID UniqueID) {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	delete(replica.excludedSegments, collectionID)
}

// addExcludedSegments will add excludedSegments to collectionReplica
func (replica *metaReplica) addExcludedSegments(collectionID UniqueID, segmentInfos []*datapb.SegmentInfo) {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	if _, ok := replica.excludedSegments[collectionID]; !ok {
		replica.excludedSegments[collectionID] = make([]*datapb.SegmentInfo, 0)
	}

	replica.excludedSegments[collectionID] = append(replica.excludedSegments[collectionID], segmentInfos...)
}

// getExcludedSegments returns excludedSegments of collectionReplica
func (replica *metaReplica) getExcludedSegments(collectionID UniqueID) ([]*datapb.SegmentInfo, error) {
	replica.mu.RLock()
	defer replica.mu.RUnlock()

	if _, ok := replica.excludedSegments[collectionID]; !ok {
		return nil, errors.New("getExcludedSegments failed, cannot found collection, id =" + fmt.Sprintln(collectionID))
	}

	return replica.excludedSegments[collectionID], nil
}

// freeAll will free all meta info from collectionReplica
func (replica *metaReplica) freeAll() {
	replica.mu.Lock()
	defer replica.mu.Unlock()

	for id := range replica.collections {
		_ = replica.removeCollectionPrivate(id)
	}

	replica.collections = make(map[UniqueID]*Collection)
	replica.partitions = make(map[UniqueID]*Partition)
	replica.segments = make(map[UniqueID]*Segment)
}

// newCollectionReplica returns a new ReplicaInterface
func newCollectionReplica(etcdKv *etcdkv.EtcdKV) ReplicaInterface {
	collections := make(map[UniqueID]*Collection)
	partitions := make(map[UniqueID]*Partition)
	segments := make(map[UniqueID]*Segment)
	excludedSegments := make(map[UniqueID][]*datapb.SegmentInfo)

	var replica ReplicaInterface = &metaReplica{
		collections: collections,
		partitions:  partitions,
		segments:    segments,

		excludedSegments: excludedSegments,
		etcdKV:           etcdKv,
	}

	return replica
}

// trans segment to queryPb.segmentInfo
func (replica *metaReplica) getSegmentInfo(segment *Segment) *querypb.SegmentInfo {
	var indexName string
	var indexID int64
	var indexInfos []*querypb.FieldIndexInfo
	// TODO:: segment has multi vec column
	indexedFieldIDs, _ := replica.getIndexedFieldIDByCollectionIDPrivate(segment.collectionID, segment)
	for _, fieldID := range indexedFieldIDs {
		fieldInfo, err := segment.getIndexedFieldInfo(fieldID)
		if err == nil {
			indexName = fieldInfo.indexInfo.IndexName
			indexID = fieldInfo.indexInfo.IndexID
			indexInfos = append(indexInfos, fieldInfo.indexInfo)
		}
	}
	info := &querypb.SegmentInfo{
		SegmentID:    segment.ID(),
		CollectionID: segment.collectionID,
		PartitionID:  segment.partitionID,
		NodeID:       Params.QueryNodeCfg.GetNodeID(),
		MemSize:      segment.getMemSize(),
		NumRows:      segment.getRowCount(),
		IndexName:    indexName,
		IndexID:      indexID,
		DmChannel:    segment.vChannelID,
		SegmentState: segment.segmentType,
		IndexInfos:   indexInfos,
		NodeIds:      []UniqueID{Params.QueryNodeCfg.GetNodeID()},
	}
	return info
}