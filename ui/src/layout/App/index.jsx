import React, { useState } from 'react';
import { Route, Routes, BrowserRouter, Outlet, useNavigate, useMatch, useLocation } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconTemplates from 'components/Icon/IconTemplates';
import Notification from 'components/Notification';
import { TooltipWrapper } from 'components/Tooltip';
import Title from 'components/Title';
import { NotificationProvider, useNotificationState, useNotificationDispatch, removeNotification } from 'context/NotificationProvider';
import { FiltersProvider, useFilterDispatch, resetFilters, resetAllFilters, FILTER_TYPES } from 'context/FiltersProvider';
import { ROUTES } from 'utils/systemConsts';
import Dashboard from 'layout/Dashboard';
import Scans from 'layout/Scans';
import Findings from 'layout/Findings';
import Assets from 'layout/Assets';
import AssetScans from 'layout/AssetScans';

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
        path: ROUTES.ASSETS,
        component: Assets,
        icon: ICON_NAMES.ASSETS,
        title: "Assets",
        resetFilters: [FILTER_TYPES.ASSETS]
    },
    {
        path: ROUTES.ASSET_SCANS,
        component: AssetScans,
        icon: ICON_NAMES.ASSET_SCANS,
        title: "Asset scans",
        resetFilters: [FILTER_TYPES.ASSET_SCANS]
    },
    {
        path: ROUTES.FINDINGS,
        component: Findings,
        icon: ICON_NAMES.FINDINGS,
        title: "Findings",
        resetFilters: [
            FILTER_TYPES.FINDINGS_GENERAL,
            FILTER_TYPES.FINDINGS_VULNERABILITIES,
            FILTER_TYPES.FINDINGS_EXPLOITS,
            FILTER_TYPES.FINDINGS_MISCONFIGURATIONS,
            FILTER_TYPES.FINDINGS_SECRETS,
            FILTER_TYPES.FINDINGS_MALWARE,
            FILTER_TYPES.FINDINGS_ROOTKITS,
            FILTER_TYPES.FINDINGS_PACKAGES
        ]
    },
    {
        path: ROUTES.SCANS,
        component: Scans,
        icon: ICON_NAMES.SCANS,
        title: "Scans",
        resetFilters: [FILTER_TYPES.SCANS, FILTER_TYPES.SCAN_CONFIGURATIONS]
    }
];

const ConnectedNotification = () => {
    const { message, type } = useNotificationState();
    const dispatch = useNotificationDispatch()

    if (!message) {
        return null;
    }

    return <Notification message={message} type={type} onClose={() => removeNotification(dispatch)} />
}

const NavLinkItem = ({ pathname, icon, resetFilterNames, resetFilterAll = false }) => {
    const location = useLocation();
    const match = useMatch(`${pathname}/*`);
    const isActive = pathname === location.pathname ? true : !!match;

    const navigate = useNavigate();
    const filtersDispatch = useFilterDispatch();

    const onClick = () => {
        if (resetFilterAll) {
            resetAllFilters(filtersDispatch);
        } else if (!!resetFilterNames) {
            resetFilters(filtersDispatch, resetFilterNames);
        }

        navigate(pathname);
    }

    return (
        <div className={classnames("nav-item", { active: isActive })} onClick={onClick}>
            <Icon name={icon} />
        </div>
    )
}

const Layout = () => {
    const { pathname } = useLocation();
    const mainPath = pathname.split("/").find(item => !!item);
    const pageTitle = ROUTES_CONFIG.find(({ path, isIndex }) => (isIndex && !mainPath) || path === `/${mainPath}`)?.title;

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date.now());

    return (
        <div id="main-wrapper">
            <FiltersProvider>
                <div className="topbar-container">
                    <img src={brandImage} alt="VMClarity" />
                    {!!pageTitle &&
                        <div className="topbar-page-title">
                            <Title medium removeMargin>{pageTitle}</Title>
                            <Icon name={ICON_NAMES.REFRESH} onClick={() => setRefreshTimestamp(Date.now())} />
                        </div>
                    }
                    <div className="topbar-menu-items">
                        <a href="/apidocs" target="_blank" className="topbar-api-link">API Docs</a>
                    </div>
                </div>
                <div className="sidebar-container">
                    {
                        ROUTES_CONFIG.map(({ path, icon, title, resetFilters, resetFilterAll }) => (
                            <TooltipWrapper key={path} tooltipId={`sidebar-item-tooltip-${path}`} tooltipText={title}>
                                <NavLinkItem pathname={path} icon={icon} resetFilterNames={resetFilters} resetFilterAll={resetFilterAll} />
                            </TooltipWrapper>
                        ))
                    }
                </div>
                <main role="main" key={refreshTimestamp}>
                    <NotificationProvider>
                        <Outlet />
                        <ConnectedNotification />
                    </NotificationProvider>
                </main>
            </FiltersProvider>
        </div>
    )
}

const App = () => (
    <div className="app-wrapper">
        <ErrorBoundary fallback={<div>Something went wrong</div>}>
            <BrowserRouter>
                <Routes>
                    <Route path="/" element={<Layout />}>
                        {
                            ROUTES_CONFIG.map(({ path, component: Component, isIndex }) => (
                                <Route key={path} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={(<Component />)} />
                            ))
                        }
                    </Route>
                </Routes>
            </BrowserRouter>
            <IconTemplates />
        </ErrorBoundary>
    </div>
)

export default App;
