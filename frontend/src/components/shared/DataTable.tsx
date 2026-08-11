import * as React from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Skeleton } from '@/components/ui/skeleton';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, Plus, Download } from 'lucide-react';

export interface DataTableColumn {
  key: string;
  label: string;
  className?: string;
}

export interface DataTableProps<T> {
  columns: DataTableColumn[];
  data: T[];
  loading?: boolean;
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  onAdd?: () => void;
  addLabel?: string;
  addSlot?: React.ReactNode;
  filters?: React.ReactNode;
  onExport?: () => void;
  emptyMessage?: string;
  renderRow: (item: T, index: number) => React.ReactNode;
  testId?: string;
}

export function DataTable<T>({ columns, data, loading, searchValue, onSearchChange, onAdd, addLabel, addSlot, filters, onExport, emptyMessage, renderRow, testId }: DataTableProps<T>) {
  return (
    <div className="space-y-3">
      {/* Filter bar */}
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between" data-testid="filter-bar">
        <div className="flex flex-1 flex-col gap-2 sm:flex-row sm:items-center">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
            <Input
              data-testid="masterdata-search-input"
              placeholder="Cerca..."
              value={searchValue || ''}
              onChange={(e) => onSearchChange?.(e.target.value)}
              className="pl-9 h-9 text-sm"
            />
          </div>
          {filters}
        </div>
        <div className="flex gap-2">
          {addSlot}
          {onExport && (
            <Button variant="outline" size="sm" onClick={onExport} className="text-xs gap-1.5">
              <Download className="h-3.5 w-3.5" /> Esporta
            </Button>
          )}
          {onAdd && (
            <Button size="sm" onClick={onAdd} className="text-xs gap-1.5" data-testid="masterdata-new-button">
              <Plus className="h-3.5 w-3.5" /> {addLabel || 'Nuovo'}
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <Card className="rounded-xl border shadow-sm" data-testid={testId || 'data-table'}>
        <div className="overflow-x-auto">
          <Table className="text-xs md:text-sm">
            <TableHeader>
              <TableRow>
                {columns.map((col) => (
                  <TableHead key={col.key} className={`py-2 text-xs font-medium ${col.className || ''}`}>{col.label}</TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={`skel-row-${i}`}>
                    {columns.map((col) => (
                      <TableCell key={`skel-${col.key}`} className="py-2"><Skeleton className="h-4 w-full" /></TableCell>
                    ))}
                  </TableRow>
                ))
              ) : data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={columns.length} className="text-center py-8 text-muted-foreground text-sm">
                    {emptyMessage || 'Nessun risultato. Modifica i filtri o crea un nuovo record.'}
                  </TableCell>
                </TableRow>
              ) : (
                data.map((item, idx) => renderRow(item, idx))
              )}
            </TableBody>
          </Table>
        </div>
      </Card>
    </div>
  );
}
