import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import clsx from 'clsx';

import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import smallLogo from '@/assets/ico/logomark.svg';

import { Link } from '@@/Link';

import fullLogoBE from './k8or-logo-v0-0-03-fil.svg';
import fullLogoCE from './k8or-logo-v0-0-03-fil.svg';
import { useSidebarState } from './useSidebarState';
import styles from './Header.module.css';

interface Props {
  logo?: string;
}

export function Header({ logo: customLogo }: Props) {
  const { toggle, isOpen } = useSidebarState();

  return (
    <div className="flex">
      <div>
        <Link
          to="portainer.home"
          data-cy="portainerSidebar-homeImage"
          className="text-2xl text-white no-underline hover:text-white hover:no-underline focus:text-white focus:no-underline focus:outline-none"
        >
          <Logo customLogo={customLogo} isOpen={isOpen} />
        </Link>
      </div>

      <button
        type="button"
        onClick={() => toggle()}
        className={clsx(
          styles.collapseBtn,
          'flex h-6 w-6 items-center justify-center rounded border-0',
          'transition-all duration-200',
          'text-sm text-gray-4 hover:text-white be:text-gray-5 be:hover:text-white',
          'bg-blue-11 be:bg-gray-10',
          'th-dark:bg-gray-warm-11',
          'absolute',
          { '-right-[10px]': !isOpen, 'right-6': isOpen }
        )}
        aria-label="Toggle Sidebar"
        title="Toggle Sidebar"
      >
        {isOpen ? <ChevronsLeft /> : <ChevronsRight />}
      </button>
    </div>
  );
}

function getLogo(isOpen: boolean, customLogo?: string) {

  return fullLogoBE;
  
  // if (customLogo) {
  //   return customLogo;
  // }

  // if (!isOpen) {
  //   return smallLogo;
  // }

  // return isBE ? fullLogoBE : fullLogoCE;
}

function Logo({
  customLogo,
  isOpen,
}: {
  customLogo?: string;
  isOpen: boolean;
}) {
  const logo = getLogo(isOpen, customLogo);


  return (
    <img
      src={logo}
      className={clsx('img-responsive', styles.logo)}
      alt="Logo"
    />
  );
}
