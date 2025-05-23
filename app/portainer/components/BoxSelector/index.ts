import angular from 'angular';

import { react2angular } from '@/react-tools/react2angular';

import { BoxSelector } from '@@/BoxSelector';

import { BoxSelectorAngular } from './BoxSelectorAngular';

export { buildOption } from './utils';
const BoxSelectorReact = react2angular(BoxSelector, [
  'isMulti',
  'value',
  'onChange',
  'options',
  'radioName',
  'slim',
  'hiddenSpacingCount',
  'error',
]);

export const boxSelectorModule = angular
  .module('app.portainer.component.box-selector', [])
  .component('boxSelectorReact', BoxSelectorReact)
  .component('boxSelector', BoxSelectorAngular).name;
