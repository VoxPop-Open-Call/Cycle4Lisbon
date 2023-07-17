import React, { useEffect, useState } from "react";

import { CSmartTable } from "@coreui/react-pro";
import { ScopedColumns } from "@coreui/react-pro/src/components/smart-table/types";

import styles from "./tableComponent.module.scss";

interface TableComponentProps<T> {
  tableData: T[];
  tableColumnProps?: ScopedColumns;
  tableColumnsVisibility: Array<{ key: string }>;
  pagination: { limit: number; offset: number; orderBy: string };
  setPagination: React.Dispatch<
    React.SetStateAction<{
      limit: number;
      offset: number;
      orderBy: string;
      type?: string;
    }>
  >;
}

const TableComponent = <T extends { id: string }>({
  tableData,
  tableColumnProps,
  tableColumnsVisibility,
  pagination,
  setPagination,
}: TableComponentProps<T>): JSX.Element => {
  const [pages, setPages] = useState({ pages: 1, limit: pagination.limit });
  useEffect(() => {
    setPages((state) => {
      const res =
        Math.ceil(pagination.offset / pagination.limit) +
        (tableData.length < pagination.limit ? 1 : 2);
      if (state.limit === pagination.limit) {
        return res > state.pages ? { ...state, pages: res } : state;
      }
      return { limit: pagination.limit, pages: res };
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tableData]);
  return (
    <CSmartTable
      items={tableData}
      columns={tableColumnsVisibility}
      clickableRows
      columnFilter
      itemsPerPageSelect
      itemsPerPage={pagination.limit}
      pagination={{ external: true }}
      scopedColumns={tableColumnProps}
      columnSorter={{ external: true }}
      //onChange methods
      onItemsPerPageChange={(itemsPerPage) => {
        setPagination((currentState) => {
          if (currentState) {
            return { ...currentState, limit: itemsPerPage };
          }
          return currentState;
        });
      }}
      onActivePageChange={(activePage) => {
        setPagination((currentState) => {
          if (currentState) {
            return {
              ...currentState,
              offset: currentState.limit * (activePage - 1),
            };
          }
          return currentState;
        });
      }}
      onSorterChange={(sorter) => {
        setPagination((currentState) => {
          if (currentState && sorter.column) {
            return {
              ...currentState,
              orderBy: `${sorter.column} ${sorter.state}`,
            };
          }
          return currentState;
        });
      }}
      //props
      paginationProps={{
        className: styles.paginationStyle,
        activePage: Math.ceil(pagination.offset / pagination.limit) + 1,
        pages: pages.pages,
      }}
      tableProps={{ responsive: true }}
    />
  );
};

export default TableComponent;
