import { RichHistoryQuery } from '../../types';
import { omit } from 'lodash';
import { SortOrder } from '../utils/richHistoryTypes';

/**
 * Temporary place for local storage specific items that are still in use in richHistory.ts
 *
 * Should be migrated to RichHistoryLocalStorage.ts
 */

export const createRetentionPeriodBoundary = (days: number, isLastTs: boolean) => {
  const today = new Date();
  const date = new Date(today.setDate(today.getDate() - days));
  /*
   * As a retention period boundaries, we consider:
   * - The last timestamp equals to the 24:00 of the last day of retention
   * - The first timestamp that equals to the 00:00 of the first day of retention
   */
  const boundary = isLastTs ? date.setHours(24, 0, 0, 0) : date.setHours(0, 0, 0, 0);
  return boundary;
};

export function filterQueriesByTime(queries: RichHistoryQuery[], timeFilter: [number, number]) {
  const filter1 = createRetentionPeriodBoundary(timeFilter[0], true); // probably the vars should have a different name
  const filter2 = createRetentionPeriodBoundary(timeFilter[1], false);
  return queries.filter((q) => q.ts < filter1 && q.ts > filter2);
}

export function filterQueriesByDataSource(queries: RichHistoryQuery[], listOfDatasourceFilters: string[]) {
  return listOfDatasourceFilters.length > 0
    ? queries.filter((q) => listOfDatasourceFilters.includes(q.datasourceName))
    : queries;
}

export function filterQueriesBySearchFilter(queries: RichHistoryQuery[], searchFilter: string) {
  return queries.filter((query) => {
    if (query.comment.includes(searchFilter)) {
      return true;
    }

    const listOfMatchingQueries = query.queries.filter((query) =>
      // Remove fields in which we don't want to be searching
      Object.values(omit(query, ['datasource', 'key', 'refId', 'hide', 'queryType'])).some((value: any) =>
        value?.toString().includes(searchFilter)
      )
    );

    return listOfMatchingQueries.length > 0;
  });
}

export const sortQueries = (array: RichHistoryQuery[], sortOrder: SortOrder) => {
  let sortFunc;

  if (sortOrder === SortOrder.Ascending) {
    sortFunc = (a: RichHistoryQuery, b: RichHistoryQuery) => (a.ts < b.ts ? -1 : a.ts > b.ts ? 1 : 0);
  }
  if (sortOrder === SortOrder.Descending) {
    sortFunc = (a: RichHistoryQuery, b: RichHistoryQuery) => (a.ts < b.ts ? 1 : a.ts > b.ts ? -1 : 0);
  }

  if (sortOrder === SortOrder.DatasourceZA) {
    sortFunc = (a: RichHistoryQuery, b: RichHistoryQuery) =>
      a.datasourceName < b.datasourceName ? -1 : a.datasourceName > b.datasourceName ? 1 : 0;
  }

  if (sortOrder === SortOrder.DatasourceAZ) {
    sortFunc = (a: RichHistoryQuery, b: RichHistoryQuery) =>
      a.datasourceName < b.datasourceName ? 1 : a.datasourceName > b.datasourceName ? -1 : 0;
  }

  return array.sort(sortFunc);
};

export const RICH_HISTORY_SETTING_KEYS = {
  retentionPeriod: 'grafana.explore.richHistory.retentionPeriod',
  starredTabAsFirstTab: 'grafana.explore.richHistory.starredTabAsFirstTab',
  activeDatasourceOnly: 'grafana.explore.richHistory.activeDatasourceOnly',
  datasourceFilters: 'grafana.explore.richHistory.datasourceFilters',
};
