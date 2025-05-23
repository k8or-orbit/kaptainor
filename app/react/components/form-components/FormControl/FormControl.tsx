import { PropsWithChildren, ReactNode } from 'react';
import clsx from 'clsx';

import { Tooltip } from '@@/Tip/Tooltip';
import { InlineLoader } from '@@/InlineLoader';

import { FormError } from '../FormError';

export type Size = 'xsmall' | 'small' | 'medium' | 'large' | 'vertical';

export interface Props {
  inputId?: string;
  dataCy?: string;
  label: ReactNode;
  size?: Size;
  tooltip?: ReactNode;
  setTooltipHtmlMessage?: boolean;
  children: ReactNode;
  errors?: ReactNode;
  required?: boolean;
  className?: string;
  isLoading?: boolean; // whether to show an inline loader, instead of the children
  loadingText?: ReactNode; // text to show when isLoading is true
}

export function FormControl({
  inputId,
  dataCy,
  label,
  size = 'small',
  tooltip = '',
  children,
  errors,
  className,
  required = false,
  setTooltipHtmlMessage,
  isLoading = false,
  loadingText = 'Loading...',
}: PropsWithChildren<Props>) {
  return (
    <div
      className={clsx(
        className,
        'form-group',
        'after:clear-both after:table after:content-[""]' // to fix issues with float
      )}
      data-cy={dataCy}
    >
      <label
        htmlFor={inputId}
        className={clsx(sizeClassLabel(size), 'control-label', 'text-left')}
      >
        {label}

        {required && <span className="text-danger">*</span>}

        {tooltip && (
          <Tooltip message={tooltip} setHtmlMessage={setTooltipHtmlMessage} />
        )}
      </label>

      <div className={clsx('flex flex-col', sizeClassChildren(size))}>
        {isLoading && (
          // 34px height to reduce layout shift when loading is complete
          <div className="h-[34px] flex items-center">
            <InlineLoader>{loadingText}</InlineLoader>
          </div>
        )}
        {!isLoading && children}
        {!!errors && !isLoading && <FormError>{errors}</FormError>}
      </div>
    </div>
  );
}

function sizeClassLabel(size?: Size) {
  switch (size) {
    case 'large':
      return 'col-sm-5 col-lg-4';
    case 'medium':
      return 'col-sm-4 col-lg-3';
    case 'xsmall':
      return 'col-sm-1';
    case 'vertical':
      return '';
    default:
      return 'col-sm-3 col-lg-2';
  }
}

function sizeClassChildren(size?: Size) {
  switch (size) {
    case 'large':
      return 'col-sm-7 col-lg-8';
    case 'medium':
      return 'col-sm-8 col-lg-9';
    case 'xsmall':
      return 'col-sm-11';
    case 'vertical':
      return '';
    default:
      return 'col-sm-9 col-lg-10';
  }
}
