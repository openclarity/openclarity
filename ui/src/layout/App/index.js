import React from 'react';
import { Route, Routes, BrowserRouter, Outlet, useNavigate, useMatch, useLocation} from 'react-router-dom';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconTemplates from 'components/Icon/IconTemplates';
import Notification from 'components/Notification';
import Tooltip from 'components/Tooltip';
import TopBarTitle from 'components/TopBarTitle';
import Dashboard from 'layout/Dashboard';
import Applications from 'layout/Applications';
import ApplicationResources from 'layout/ApplicationResources';
import Packages from 'layout/Packages';
import Vulnerabilities from 'layout/Vulnerabilities';
import RuntimeScan from 'layout/RuntimeScan';
import { NotificationProvider, useNotificationState, useNotificationDispatch, removeNotification } from 'context/NotificationProvider';
import { FiltersProvider, useFilterDispatch, resetFilters, resetAllFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { ROUTES } from 'utils/systemConsts';

import brandImage from 'utils/images/brand.svg';

import './app.scss';

const ROUTES_CONFIG = [
    {
		path: ROUTES.DEFAULT,
		component: Dashboard,
        icon: ICON_NAMES.DASHBOARD,
        isIndex: true,
        title: "Dashboard",
        resetFilterAll: true
	},
	{
		path: ROUTES.APPLICATIONS,
		component: Applications,
        icon: ICON_NAMES.APPLICATION,
        resetFilter: FILTER_TYPES.APPLICATIONS,
        title: "Applications"
	},
    {
		path: ROUTES.APPLICATION_RESOURCES,
		component: ApplicationResources,
        icon: ICON_NAMES.RESOURCE,
        resetFilter: FILTER_TYPES.APPLICATION_RESOURCES,
        title: "Applications Resources"
	},
    {
		path: ROUTES.PACKAGES,
		component: Packages,
        icon: ICON_NAMES.PACKAGE,
        resetFilter: FILTER_TYPES.PACKAGES,
        title: "Packages"
	},
    {
		path: ROUTES.VULNERABILITIES,
		component: Vulnerabilities,
        icon: ICON_NAMES.VULNERABILITY,
        resetFilter: FILTER_TYPES.VULNERABILITIES,
        title: "Vulnerabilities"
	},
    {
		path: ROUTES.RUNTIME_SCAN,
		component: RuntimeScan,
        icon: ICON_NAMES.SCAN,
        title: "Runtime Scan"
	}
];

const ConnectedNotification = () => {
    const {message, type} = useNotificationState();
    const dispatch = useNotificationDispatch()

    if (!message) {
        return null;
    }

    return <Notification message={message} type={type} onClose={() => removeNotification(dispatch)} />
}

const NavLinkItem = ({pathname, icon, resetFilter, resetFilterAll=false}) => {
    const location = useLocation();
    const match = useMatch(`${pathname}/*`);
    const isActive = pathname === location.pathname ? true : !!match;

    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        if (resetFilterAll) {
            resetAllFilters(filtersDispatch);
        } else {
            resetFilters(filtersDispatch, resetFilter);
        }

        navigate(pathname);
    }
    
    return (
        <div className={classnames("nav-item", {active: isActive})} onClick={onClick}>
            <Icon name={icon} />
        </div>
    )
}

const Layout = () => (
    <div id="main-wrapper">
        <TopBarTitle />
        <FiltersProvider>
            <div className="sidebar-container">
                <div className="app-logo-wrapper">
                    <img src={brandImage} alt="KubeClarity" />
                </div>
                {
                    ROUTES_CONFIG.map(({path, icon, resetFilter, resetFilterAll, title}) => {
                        const tooltipId = `sidebar-item-tooltip-${path}`;

                        return (
                            <div key={path} data-tip data-for={tooltipId}>
                                <NavLinkItem pathname={path} icon={icon} resetFilter={resetFilter} resetFilterAll={resetFilterAll} />
                                <Tooltip id={tooltipId} text={title} />
                            </div>
                        )
                    })
                }
            </div>
            <main role="main">
                <NotificationProvider>
                    <Outlet />
                    <ConnectedNotification />
                </NotificationProvider>
            </main>
        </FiltersProvider>
    </div>
)

const App = () => (
    <div className="app-wrapper">
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Layout />}>
                    {
                        ROUTES_CONFIG.map(({path, component: Component, isIndex}) => (
                            <Route key={path} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={<Component />} />
                        ))
                    }
                </Route>
            </Routes>
        </BrowserRouter>
        <IconTemplates />
    </div>
)

export default App;