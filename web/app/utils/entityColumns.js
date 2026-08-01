import LocalizedText from '~/components/display/LocalizedText.vue'
import StatusBadge from '~/components/display/StatusBadge.vue'
import DateLabel from '~/components/display/DateLabel.vue'
import RelationLabel from '~/components/display/RelationLabel.vue'
import RelationListLabel from '~/components/display/RelationListLabel.vue'
import EnumLabel from '~/components/display/EnumLabel.vue'
import MoneyAmountLabel from '~/components/display/MoneyAmountLabel.vue'
import QuantityAmountLabel from '~/components/display/QuantityAmountLabel.vue'
import SaleLink from '~/components/display/SaleLink.vue'
import SaleWarehousesLabel from '~/components/display/SaleWarehousesLabel.vue'
import ImageThumbnail from '~/components/display/ImageThumbnail.vue'
import { STATUS_COLOR_MAP } from '~/design/tokens.js'

// Shared list-column -> display-component mapping (TD §9.3), used by both
// EntityDataTable (list rows) and EntityViewDrawer (single-record view) so
// there's exactly one place that knows how to render a column's value.
export const COLUMN_COMPONENTS = {
  LocalizedText,
  StatusBadge,
  DateLabel,
  RelationLabel,
  RelationListLabel,
  EnumLabel,
  MoneyAmountLabel,
  QuantityAmountLabel,
  SaleLink,
  SaleWarehousesLabel,
  ImageThumbnail,
}

export function columnComponent(col) {
  return COLUMN_COMPONENTS[col.component]
}

export function columnProps(col, item) {
  if (col.component === 'StatusBadge') return { status: item[col.key], map: STATUS_COLOR_MAP[col.statusMap] }
  if (col.component === 'LocalizedText') return { value: item[col.key] }
  if (col.component === 'DateLabel') return { value: item[col.key] }
  if (col.component === 'RelationLabel') return { value: item[col.key], relation: col.relation }
  if (col.component === 'RelationListLabel') return { value: item[col.key], relation: col.relation }
  if (col.component === 'EnumLabel') return { value: item[col.key] }
  if (col.component === 'MoneyAmountLabel') return { value: item[col.key] }
  if (col.component === 'QuantityAmountLabel') return { value: item[col.key] }
  if (col.component === 'SaleLink') return { value: item[col.key] }
  if (col.component === 'SaleWarehousesLabel') return { value: item[col.key] }
  if (col.component === 'ImageThumbnail') return { value: item[col.key] }
  return {}
}
