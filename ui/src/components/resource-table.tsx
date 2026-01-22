import React, { useCallback, useEffect, useMemo, useState } from 'react'
import {
  ColumnDef,
  ColumnFiltersState,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  PaginationState,
  RowSelectionState,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import {
  Box,
  Database,
  Plus,
  RefreshCw,
  Search,
  Settings2,
  Trash2,
  XCircle,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ResourceType } from '@/types/api'
import { deleteResource, useResources, useResourcesWatch } from '@/lib/api'
import { useCluster } from '@/hooks/use-cluster'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

import { ConnectionIndicator } from './connection-indicator'
import { ErrorMessage } from './error-message'
import { ResourceTableView } from './resource-table-view'
import { NamespaceSelector } from './selector/namespace-selector'

export interface ResourceTableProps<T> {
  resourceName: string
  resourceType?: ResourceType // Optional, used for fetching resources
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<T, any>[]
  clusterScope?: boolean // If true, don't show namespace selector
  searchQueryFilter?: (item: T, query: string) => boolean // Custom filter function
  showCreateButton?: boolean // If true, show create button
  onCreateClick?: () => void // Callback for create button click
  extraToolbars?: React.ReactNode[] // Additional toolbar components
  defaultHiddenColumns?: string[] // Columns to hide by default
  data?: T[] // External data (skips internal fetching if provided)
  isLoading?: boolean // Loading state for external data
  hideFilter?: boolean // Hide filter controls
  disablePagination?: boolean // Disable pagination controls
}

export function ResourceTable<T>({
  resourceName,
  resourceType,
  columns,
  clusterScope = false,
  searchQueryFilter,
  showCreateButton = false,
  onCreateClick,
  extraToolbars = [],
  defaultHiddenColumns = [],
  data: externalData,
  isLoading: externalLoading = false,
  hideFilter = false,
  disablePagination = false,
}: ResourceTableProps<T>) {
  const { t } = useTranslation()
  const { currentCluster } = useCluster()
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-columnFilters`
    const savedFilters = sessionStorage.getItem(storageKey)
    return savedFilters ? JSON.parse(savedFilters) : []
  })
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [searchQuery, setSearchQuery] = useState<string>(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-searchQuery`
    return sessionStorage.getItem(storageKey) || ''
  })

  const [columnVisibility, setColumnVisibility] = useState<
    Record<string, boolean>
  >(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-columnVisibility`
    const savedVisibility = localStorage.getItem(storageKey)
    if (savedVisibility) {
      return JSON.parse(savedVisibility)
    }
    // Set default hidden columns if no saved state
    const initialVisibility: Record<string, boolean> = {}
    defaultHiddenColumns.forEach((colId) => {
      initialVisibility[colId] = false
    })
    return initialVisibility
  })

  const [pagination, setPagination] = useState<PaginationState>(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-pageSize`
    const savedPageSize = sessionStorage.getItem(storageKey)
    return {
      pageIndex: 0,
      pageSize: savedPageSize ? Number(savedPageSize) : 20,
    }
  })
  const [refreshInterval, setRefreshInterval] = useState(5000)

  const [selectedNamespace, setSelectedNamespace] = useState<
    string | undefined
  >(() => {
    // Try to get the stored namespace from localStorage
    const storedNamespace = localStorage.getItem(
      localStorage.getItem('current-cluster') + 'selectedNamespace'
    )
    return clusterScope
      ? undefined // No namespace for cluster scope
      : storedNamespace || '_all' // Default to '_all' if not set
  })
  const [useSSE, setUseSSE] = useState(false)
  const {
    isLoading: queryLoading,
    data: queryData,
    isError: queryIsError,
    error: queryError,
    refetch: queryRefetch,
  } = useResources(
    resourceType ?? (resourceName.toLowerCase() as ResourceType),
    selectedNamespace,
    {
      refreshInterval: useSSE ? 0 : refreshInterval, // disable polling when SSE
      reduce: true, // Fetch reduced data for performance
      disable: useSSE || !!externalData, // do not query when using SSE or external data provided
    }
  )

  // SSE state (when enabled)
  // SSE watch hook
  const {
    data: watchData,
    isLoading: watchLoading,
    error: watchError,
    isConnected,
    refetch: reconnectSSE,
  } = useResourcesWatch(
    (resourceType ??
      (resourceName.toLowerCase() as ResourceType)) as ResourceType,
    selectedNamespace,
    {
      reduce: true,
      enabled: useSSE,
      clusterName: currentCluster || undefined, // Pass cluster from URL context
    }
  )

  // (moved below after error is defined)

  // Update sessionStorage when search query changes
  useEffect(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-searchQuery`
    if (searchQuery) {
      sessionStorage.setItem(storageKey, searchQuery)
    } else {
      sessionStorage.removeItem(storageKey)
    }
  }, [searchQuery, resourceName])

  // Update sessionStorage when column visibility changes
  useEffect(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-columnVisibility`
    localStorage.setItem(storageKey, JSON.stringify(columnVisibility))
  }, [columnVisibility, resourceName])

  // Update sessionStorage when page size changes
  useEffect(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-pageSize`
    sessionStorage.setItem(storageKey, pagination.pageSize.toString())
  }, [pagination.pageSize, resourceName])

  // Update sessionStorage when column filters changes
  useEffect(() => {
    const currentCluster = localStorage.getItem('current-cluster')
    const storageKey = `${currentCluster}-${resourceName}-columnFilters`
    if (columnFilters.length > 0) {
      sessionStorage.setItem(storageKey, JSON.stringify(columnFilters))
    } else {
      sessionStorage.removeItem(storageKey)
    }
  }, [columnFilters, resourceName])

  // Reset pagination when filters change
  useEffect(() => {
    setPagination((prev) => ({ ...prev, pageIndex: 0 }))
  }, [columnFilters, searchQuery])

  // Handle namespace change
  const handleNamespaceChange = useCallback(
    (value: string) => {
      if (setSelectedNamespace) {
        localStorage.setItem(
          localStorage.getItem('current-cluster') + 'selectedNamespace',
          value
        )
        setSelectedNamespace(value)
        // Reset pagination and search when changing namespace
        setPagination({ pageIndex: 0, pageSize: pagination.pageSize })
        setSearchQuery('')
      }
    },
    [setSelectedNamespace, pagination.pageSize]
  )

  // Add namespace column when showing all namespaces
  const enhancedColumns = useMemo(() => {
    const selectColumn: ColumnDef<T> = {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    }

    const baseColumns = [selectColumn, ...columns]

    const isAllNamespaces = selectedNamespace === '_all' || !selectedNamespace
    const isMultiNamespace = selectedNamespace?.includes(',')

    // Only add namespace column if not cluster scope, showing multiple or all namespaces,
    // and there isn't already a namespace column in the provided columns
    if (!clusterScope && (isAllNamespaces || isMultiNamespace)) {
      // Check if namespace column already exists in the provided columns
      const hasNamespaceColumn = columns.some((col) => {
        // Check if the column accesses namespace data
        if ('accessorKey' in col && col.accessorKey === 'metadata.namespace') {
          return true
        }
        if ('accessorFn' in col && col.id === 'namespace') {
          return true
        }
        return false
      })

      // Only add namespace column if it doesn't already exist
      if (!hasNamespaceColumn) {
        const namespaceColumn = {
          id: 'namespace',
          header: t('resourceTable.namespace'),
          accessorFn: (row: T) => {
            // Try to get namespace from metadata.namespace
            const metadata = (row as { metadata?: { namespace?: string } })
              ?.metadata
            // Fallback for resources like HelmRelease
            if (!metadata) {
              const flatNamespace = (row as { namespace?: string })?.namespace
              return flatNamespace || '-'
            }
            return metadata?.namespace || '-'
          },
          cell: ({ getValue }: { getValue: () => string }) => (
            <Badge variant="outline" className="ml-2 ">
              {getValue()}
            </Badge>
          ),
        }

        // Insert namespace column after select and first column (typically name)
        const columnsWithNamespace = [...baseColumns]
        columnsWithNamespace.splice(2, 0, namespaceColumn)
        return columnsWithNamespace
      }
    }
    return baseColumns
  }, [columns, clusterScope, selectedNamespace, t])

  const data = useMemo(() => {
    if (externalData) return externalData
    if (useSSE) return watchData
    return queryData
  }, [useSSE, watchData, queryData, externalData])
  const isLoading = externalData
    ? (externalLoading ?? false)
    : useSSE
      ? watchLoading
      : queryLoading
  const isError = externalData
    ? false
    : useSSE
      ? Boolean(watchError)
      : queryIsError
  const error = externalData
    ? null
    : useSSE
      ? (watchError as Error | null)
      : (queryError as unknown as Error | null)
  const refetch = externalData ? () => {} : useSSE ? reconnectSSE : queryRefetch

  const memoizedData = useMemo(() => (data || []) as T[], [data])

  useEffect(() => {
    if (!useSSE && error) {
      setRefreshInterval(0)
    }
  }, [useSSE, error])

  // Create table instance using TanStack Table
  const table = useReactTable<T>({
    data: memoizedData,
    columns: enhancedColumns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onRowSelectionChange: setRowSelection,
    onColumnVisibilityChange: setColumnVisibility,
    getRowId: (row) => {
      const metadata = (
        row as {
          metadata?: { name?: string; namespace?: string; uid?: string }
        }
      )?.metadata

      if (!metadata) {
        // Fallback for HelmRelease
        const flat = row as { name?: string; namespace?: string }
        if (flat.name) {
          return flat.namespace ? `${flat.namespace}/${flat.name}` : flat.name
        }
      }

      if (!metadata?.name) {
        return `row-${Math.random()}`
      }
      return (
        metadata.uid ||
        (metadata.namespace
          ? `${metadata.namespace}/${metadata.name}`
          : metadata.name)
      )
    },
    state: {
      sorting,
      columnFilters,
      globalFilter: searchQuery,
      pagination,
      rowSelection,
      columnVisibility,
    },
    onPaginationChange: setPagination,
    // Let TanStack Table handle pagination automatically based on filtered data
    manualPagination: false,
    // Improve filtering performance and consistency
    globalFilterFn: (row, _columnId, value) => {
      if (searchQueryFilter) {
        return searchQueryFilter(row.original as T, String(value).toLowerCase())
      }
      const searchValue = String(value).toLowerCase()

      // Search across all visible columns
      return row.getVisibleCells().some((cell) => {
        const cellValue = String(cell.getValue() || '').toLowerCase()
        return cellValue.includes(searchValue)
      })
    },
    // Add this to prevent unnecessary pagination resets
    autoResetPageIndex: false,
    enableRowSelection: true,
  })

  // Handle batch delete - must be after table is defined
  const handleBatchDelete = useCallback(async () => {
    setIsDeleting(true)
    const selectedRows = table
      .getSelectedRowModel()
      .rows.map((row) => row.original)

    const deletePromises = selectedRows.map((row) => {
      const metadata = (
        row as { metadata?: { name?: string; namespace?: string } }
      )?.metadata

      let name = metadata?.name
      let namespace = clusterScope ? undefined : metadata?.namespace

      if (!metadata) {
        const flat = row as { name?: string; namespace?: string }
        name = flat.name
        if (!clusterScope) {
          namespace = flat.namespace
        }
      }

      if (!name) {
        return Promise.resolve()
      }

      return deleteResource(
        resourceType ?? (resourceName.toLowerCase() as ResourceType),
        name,
        namespace
      )
        .then(() => {
          toast.success(t('resourceTable.deleteSuccess', { name }))
        })
        .catch((error) => {
          console.error(`Failed to delete ${name}:`, error)
          toast.error(
            t('resourceTable.deleteFailed', { name, error: error.message })
          )
          throw error
        })
    })

    try {
      await Promise.allSettled(deletePromises)
      // Reset selection and close dialog
      setRowSelection({})
      setDeleteDialogOpen(false)
      // Refetch data
      if (!useSSE) {
        refetch()
      }
    } finally {
      setIsDeleting(false)
    }
  }, [table, clusterScope, resourceType, resourceName, t, useSSE, refetch])
  // Calculate total and filtered row counts
  const totalRowCount = useMemo(
    () => (data as T[] | undefined)?.length || 0,
    [data]
  )
  const filteredRowCount = useMemo(() => {
    if (!data || (data as T[]).length === 0) return 0
    // Force re-computation when filters change
    void searchQuery // Ensure dependency is used
    void columnFilters // Ensure dependency is used
    return table.getFilteredRowModel().rows.length
  }, [table, data, searchQuery, columnFilters])

  // Check if there are active filters
  const hasActiveFilters = useMemo(() => {
    return Boolean(searchQuery) || columnFilters.length > 0
  }, [searchQuery, columnFilters])

  // Render empty state based on condition
  const renderEmptyState = () => {
    // Only show loading state if there's no existing data
    if (isLoading && (!data || (data as T[]).length === 0)) {
      return (
        <div className="h-72 flex flex-col items-center justify-center">
          <div className="mb-4 bg-muted/30 p-6 rounded-full">
            <Database className="h-12 w-12 text-muted-foreground animate-pulse" />
          </div>
          <h3 className="text-lg font-medium mb-1">
            Loading {resourceName.toLowerCase()}...
          </h3>
          <p className="text-muted-foreground">
            Retrieving data
            {!clusterScope && selectedNamespace
              ? ` from ${selectedNamespace === '_all' ? 'All Namespaces' : `namespace ${selectedNamespace}`}`
              : ''}
          </p>
        </div>
      )
    }

    if (isError) {
      return (
        <ErrorMessage
          resourceName={resourceName}
          error={error}
          refetch={refetch}
        />
      )
    }

    if (data && (data as T[]).length === 0) {
      return (
        <div className="h-72 flex flex-col items-center justify-center">
          <div className="mb-4 bg-muted/30 p-6 rounded-full">
            <Box className="h-12 w-12 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-medium mb-1">
            No {resourceName.toLowerCase()} found
          </h3>
          <p className="text-muted-foreground">
            {searchQuery
              ? `No results match your search query: "${searchQuery}"`
              : clusterScope
                ? `There are no ${resourceName.toLowerCase()} found`
                : `There are no ${resourceName.toLowerCase()} ${selectedNamespace === '_all' ? 'in All Namespaces' : `in the ${selectedNamespace} namespace`}`}
          </p>
          {searchQuery && (
            <Button
              variant="outline"
              className="mt-4"
              onClick={() => setSearchQuery('')}
            >
              Clear Search
            </Button>
          )}
        </div>
      )
    }

    return null
  }

  const emptyState = renderEmptyState()

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold capitalize">{resourceName}</h1>
          {!clusterScope && selectedNamespace && (
            <div className="text-muted-foreground flex items-center mt-1">
              <span>Namespace:</span>
              <div className="flex flex-wrap gap-1 ml-2">
                {selectedNamespace === '_all' || !selectedNamespace ? (
                  <Badge variant="outline">All Namespaces</Badge>
                ) : (
                  selectedNamespace.split(',').map((ns) => (
                    <Badge key={ns} variant="outline">
                      {ns}
                    </Badge>
                  ))
                )}
              </div>
            </div>
          )}
        </div>

        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
          <div className="flex items-center gap-2 flex-wrap">
            {extraToolbars?.map((toolbar, index) => (
              <React.Fragment key={index}>{toolbar}</React.Fragment>
            ))}
            {/* Watch/Live mode toggle switch */}
            {resourceName === 'Pods' && (
              <div className="flex items-center gap-2">
                <Label className="text-sm">
                  {useSSE ? (
                    <ConnectionIndicator isConnected={isConnected}>
                      {t('resourceTable.watch')}
                    </ConnectionIndicator>
                  ) : (
                    t('resourceTable.watch')
                  )}
                </Label>
                <Switch
                  checked={useSSE}
                  onCheckedChange={(checked) => {
                    setUseSSE(checked)
                    if (checked) {
                      setRefreshInterval(0)
                    } else if (refreshInterval === 0) {
                      setRefreshInterval(5000) // Default to 5s when disabling watch mode
                    }
                  }}
                />
              </div>
            )}
            {/* Refresh interval selector */}
            {!hideFilter && (
              <Select
                value={refreshInterval.toString()}
                onValueChange={(value) => {
                  setRefreshInterval(Number(value))
                  if (Number(value) > 0) {
                    setUseSSE(false)
                  }
                }}
                disabled={useSSE}
              >
                <SelectTrigger className="max-w-[140px]">
                  <div className="flex items-center gap-2">
                    <RefreshCw className="h-4 w-4" />
                    <SelectValue />
                  </div>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">Off</SelectItem>
                  <SelectItem value="1000">1s</SelectItem>
                  <SelectItem value="5000">5s</SelectItem>
                  <SelectItem value="10000">10s</SelectItem>
                  <SelectItem value="30000">30s</SelectItem>
                </SelectContent>
              </Select>
            )}
            {!clusterScope && !hideFilter && (
              <NamespaceSelector
                selectedNamespace={selectedNamespace}
                handleNamespaceChange={handleNamespaceChange}
                showAll={true}
                multiple={true}
              />
            )}
            {/* Column Filters */}
            {!hideFilter &&
              table
                .getAllColumns()
                .filter((column) => {
                  const columnDef = column.columnDef as ColumnDef<T> & {
                    enableColumnFilter?: boolean
                  }
                  return columnDef.enableColumnFilter && column.getCanFilter()
                })
                .map((column) => {
                  const columnDef = column.columnDef as ColumnDef<T> & {
                    enableColumnFilter?: boolean
                  }
                  const uniqueValues = column.getFacetedUniqueValues()
                  const filterValue = column.getFilterValue() as string

                  return (
                    <Select
                      key={column.id}
                      value={filterValue || ''}
                      onValueChange={(value) =>
                        column.setFilterValue(value === 'all' ? '' : value)
                      }
                    >
                      <SelectTrigger className="min-w-32">
                        <SelectValue
                          placeholder={`Filter ${typeof columnDef.header === 'string' ? columnDef.header : 'Column'}`}
                        />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">
                          All{' '}
                          {typeof columnDef.header === 'string'
                            ? columnDef.header
                            : 'Values'}
                        </SelectItem>
                        {Array.from(uniqueValues.keys())
                          .sort()
                          .map((value) =>
                            value ? (
                              <SelectItem
                                key={String(value)}
                                value={String(value)}
                              >
                                {String(value)} ({uniqueValues.get(value)})
                              </SelectItem>
                            ) : null
                          )}
                      </SelectContent>
                    </Select>
                  )
                })}
          </div>

          {/* Search bar */}
          {!hideFilter && (
            <div className="flex items-center gap-2 w-full sm:w-auto">
              <div className="relative w-full sm:w-auto">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={`Search ${resourceName.toLowerCase()}...`}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 pr-4 w-full sm:w-[100px] md:w-[200px]"
                />
              </div>
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setSearchQuery('')}
                  className="h-9 w-9"
                >
                  <XCircle className="h-4 w-4" />
                </Button>
              )}
            </div>
          )}
          {/* Batch delete button */}
          {table.getSelectedRowModel().rows.length > 0 && (
            <Button
              variant="destructive"
              onClick={() => setDeleteDialogOpen(true)}
              className="gap-2"
            >
              <Trash2 className="h-4 w-4" />
              {t('resourceTable.deleteSelected', {
                count: table.getSelectedRowModel().rows.length,
              })}
            </Button>
          )}
          {showCreateButton && onCreateClick && (
            <Button onClick={onCreateClick} className="gap-1">
              <Plus className="h-2 w-2" />
              New
            </Button>
          )}

          {/* Toggle columns Dropdown */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="icon" className="h-9 w-9">
                <Settings2 className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Toggle columns</DropdownMenuLabel>
              <DropdownMenuSeparator />
              {table
                .getAllLeafColumns()
                .filter((column) => column.getCanHide())
                .map((column) => {
                  const header = column.columnDef.header
                  const headerText =
                    typeof header === 'string' ? header : column.id
                  return (
                    <DropdownMenuCheckboxItem
                      key={column.id}
                      className="capitalize"
                      checked={column.getIsVisible()}
                      onCheckedChange={(value) =>
                        column.toggleVisibility(!!value)
                      }
                    >
                      {headerText}
                    </DropdownMenuCheckboxItem>
                  )
                })}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      <ResourceTableView
        table={table}
        columnCount={enhancedColumns.length}
        isLoading={isLoading}
        data={data as T[] | undefined}
        emptyState={emptyState}
        hasActiveFilters={hasActiveFilters}
        filteredRowCount={filteredRowCount}
        totalRowCount={totalRowCount}
        searchQuery={searchQuery}
        pagination={pagination}
        setPagination={setPagination}
        disablePagination={disablePagination}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('resourceTable.confirmDeletion')}</DialogTitle>
            <DialogDescription>
              {t('resourceTable.confirmDeletionMessage', {
                count: table.getSelectedRowModel().rows.length,
                resourceName: resourceName.toLowerCase(),
              })}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={isDeleting}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="destructive"
              onClick={handleBatchDelete}
              disabled={isDeleting}
            >
              {isDeleting ? t('resourceTable.deleting') : t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
