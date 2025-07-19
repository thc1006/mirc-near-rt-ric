// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dataselect provides modernized data selection using Go generics
package dataselect

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
)

// ModernComparableValue represents a value that can be compared and searched
type ModernComparableValue[T any] interface {
	// Compare returns -1 if less than other, 0 if equal, 1 if greater
	Compare(other T) int
	// Contains returns true if this value contains or matches the other value
	Contains(other T) bool
	// Value returns the underlying value
	Value() T
}

// ModernDataCell represents a data cell with type safety
type ModernDataCell[T any] interface {
	// GetProperty returns a comparable property by name
	GetProperty(name PropertyName) ModernComparableValue[T]
	// GetData returns the underlying data
	GetData() T
}

// ModernMetricDataCell extends ModernDataCell with metric capabilities
type ModernMetricDataCell[T any] interface {
	ModernDataCell[T]
	// GetResourceSelector returns selector for metrics
	GetResourceSelector() *metricapi.ResourceSelector
}

// ModernDataSelector provides type-safe data selection operations
type ModernDataSelector[T any, C ModernDataCell[T]] struct {
	dataList        []C
	query          *DataSelectQuery
	cachedResources *metricapi.CachedResources
	logger         *slog.Logger
}

// NewModernDataSelector creates a new type-safe data selector
func NewModernDataSelector[T any, C ModernDataCell[T]](
	data []C,
	query *DataSelectQuery,
	cached *metricapi.CachedResources,
	logger *slog.Logger,
) *ModernDataSelector[T, C] {
	if logger == nil {
		logger = slog.Default()
	}
	
	return &ModernDataSelector[T, C]{
		dataList:        data,
		query:          query,
		cachedResources: cached,
		logger:         logger.With("component", "modern_data_selector"),
	}
}

// Len returns the number of items in the selector
func (mds *ModernDataSelector[T, C]) Len() int {
	return len(mds.dataList)
}

// Sort sorts the data according to the sort query
func (mds *ModernDataSelector[T, C]) Sort() *ModernDataSelector[T, C] {
	if mds.query == nil || mds.query.SortQuery == nil {
		return mds
	}

	start := len(mds.dataList)
	mds.logger.Debug("starting sort operation",
		"items", start,
		"sort_fields", len(mds.query.SortQuery.SortByList))

	slices.SortFunc(mds.dataList, func(a, b C) int {
		for _, sortBy := range mds.query.SortQuery.SortByList {
			propA := a.GetProperty(sortBy.Property)
			propB := b.GetProperty(sortBy.Property)

			// Skip if properties are not available
			if propA == nil || propB == nil {
				mds.logger.Warn("sort property not found",
					"property", sortBy.Property,
					"item_type", fmt.Sprintf("%T", a))
				continue
			}

			// Get comparison result
			comparison := propA.Compare(propB.Value())
			if comparison != 0 {
				// Apply sort order
				if sortBy.Ascending {
					return comparison
				}
				return -comparison
			}
		}
		return 0
	})

	mds.logger.Debug("sort operation completed",
		"items", len(mds.dataList),
		"sorted", true)

	return mds
}

// Filter applies filter criteria to the data
func (mds *ModernDataSelector[T, C]) Filter() *ModernDataSelector[T, C] {
	if mds.query == nil || mds.query.FilterQuery == nil {
		return mds
	}

	originalCount := len(mds.dataList)
	
	// Use slices.DeleteFunc for efficient filtering
	mds.dataList = slices.DeleteFunc(mds.dataList, func(item C) bool {
		return !mds.matchesFilter(item)
	})

	mds.logger.Debug("filter operation completed",
		"original_count", originalCount,
		"filtered_count", len(mds.dataList),
		"removed", originalCount-len(mds.dataList))

	return mds
}

// FilterWith applies a custom filter predicate
func (mds *ModernDataSelector[T, C]) FilterWith(predicate func(C) bool) *ModernDataSelector[T, C] {
	originalCount := len(mds.dataList)
	
	mds.dataList = slices.DeleteFunc(mds.dataList, func(item C) bool {
		return !predicate(item)
	})

	mds.logger.Debug("custom filter operation completed",
		"original_count", originalCount,
		"filtered_count", len(mds.dataList))

	return mds
}

// Paginate applies pagination to the data
func (mds *ModernDataSelector[T, C]) Paginate() *ModernDataSelector[T, C] {
	if mds.query == nil || mds.query.PaginationQuery == nil {
		return mds
	}

	pq := mds.query.PaginationQuery
	totalItems := len(mds.dataList)
	
	// Calculate pagination bounds
	start := pq.ItemsPerPage * pq.Page
	if start > totalItems {
		mds.dataList = mds.dataList[:0]
		mds.logger.Debug("pagination beyond data range",
			"page", pq.Page,
			"items_per_page", pq.ItemsPerPage,
			"total_items", totalItems,
			"result_count", 0)
		return mds
	}

	end := start + pq.ItemsPerPage
	if end > totalItems {
		end = totalItems
	}

	mds.dataList = mds.dataList[start:end]
	
	mds.logger.Debug("pagination applied",
		"page", pq.Page,
		"items_per_page", pq.ItemsPerPage,
		"total_items", totalItems,
		"result_count", len(mds.dataList),
		"start", start,
		"end", end)

	return mds
}

// GetData returns the current data list
func (mds *ModernDataSelector[T, C]) GetData() []C {
	return mds.dataList
}

// GetUnderlyingData extracts the underlying data from all cells
func (mds *ModernDataSelector[T, C]) GetUnderlyingData() []T {
	result := make([]T, len(mds.dataList))
	for i, cell := range mds.dataList {
		result[i] = cell.GetData()
	}
	return result
}

// Select performs the complete selection operation (sort, filter, paginate)
func (mds *ModernDataSelector[T, C]) Select() *ModernDataSelector[T, C] {
	return mds.Sort().Filter().Paginate()
}

// matchesFilter checks if an item matches the current filter query
func (mds *ModernDataSelector[T, C]) matchesFilter(item C) bool {
	if mds.query.FilterQuery == nil {
		return true
	}

	for _, filterBy := range mds.query.FilterQuery.FilterByList {
		property := item.GetProperty(filterBy.Property)
		if property == nil {
			mds.logger.Warn("filter property not found",
				"property", filterBy.Property,
				"item_type", fmt.Sprintf("%T", item))
			continue
		}

		// Create a comparable value for the filter value
		filterValue := NewModernStringValue(filterBy.Value)
		
		if !property.Contains(filterValue.Value()) {
			return false
		}
	}

	return true
}

// ModernStringValue implements ModernComparableValue for strings
type ModernStringValue struct {
	value string
}

// NewModernStringValue creates a new string comparable value
func NewModernStringValue(value string) ModernComparableValue[string] {
	return &ModernStringValue{value: value}
}

func (msv *ModernStringValue) Compare(other string) int {
	return cmp.Compare(msv.value, other)
}

func (msv *ModernStringValue) Contains(other string) bool {
	return msv.value == other || len(other) == 0
}

func (msv *ModernStringValue) Value() string {
	return msv.value
}

// ModernIntValue implements ModernComparableValue for integers
type ModernIntValue struct {
	value int
}

// NewModernIntValue creates a new int comparable value
func NewModernIntValue(value int) ModernComparableValue[int] {
	return &ModernIntValue{value: value}
}

func (miv *ModernIntValue) Compare(other int) int {
	return cmp.Compare(miv.value, other)
}

func (miv *ModernIntValue) Contains(other int) bool {
	return miv.value == other
}

func (miv *ModernIntValue) Value() int {
	return miv.value
}

// Helper functions for common operations

// SelectResources is a generic function for selecting Kubernetes resources
func SelectResources[T any, C ModernDataCell[T]](
	resources []C,
	query *DataSelectQuery,
	logger *slog.Logger,
) ([]C, error) {
	if len(resources) == 0 {
		return resources, nil
	}

	selector := NewModernDataSelector(resources, query, nil, logger)
	result := selector.Select()
	
	if result.Len() > MaxPageSize {
		return nil, fmt.Errorf("%w: requested %d items, maximum allowed is %d",
			errors.ErrInvalidInput, result.Len(), MaxPageSize)
	}

	return result.GetData(), nil
}

// Constants for validation
const (
	MaxPageSize = 1000
)