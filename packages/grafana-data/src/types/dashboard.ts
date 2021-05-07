import { FieldConfigSource } from './fieldOverrides';
import { DataQuery, DatasourceRef } from './query';
import { ScopedVars } from './ScopedVars';

export type ModelVersion = [number, number];

export interface PanelModel<TOptions = any> {
  /** ID of the panel within the current dashboard */
  id: number;

  /** Panel options */
  options: TOptions;

  /** Field options configuration */
  fieldConfig: FieldConfigSource;

  /** Version of the panel plugin */
  pluginVersion?: string;

  /** The model version for the plugin */
  pluginModel?: ModelVersion;

  /** The datasource used in all targets */
  datasource?: DatasourceRef | null;

  /** The queries in a panel */
  targets?: DataQuery[];

  // RUNTIME!!!! (not saved in panel model)
  scopedVars?: ScopedVars;
}
