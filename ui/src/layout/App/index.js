import React from 'react';
import { Route, Routes, BrowserRouter, Outlet, useNavigate, useMatch, useLocation} from 'react-router-dom';
import classnames from 'classnames';
import Icon, { ICON_NAMES } from 'components/Icon';
import IconTemplates from 'components/Icon/IconTemplates';
import Notification from 'components/Notification';
import { TooltipWrapper } from 'components/Tooltip';
import Title from 'components/Title';
import Scans from 'layout/Scans';
import Findings from 'layout/Findings';
import { NotificationProvider, useNotificationState, useNotificationDispatch, removeNotification } from 'context/NotificationProvider';
import { ROUTES } from 'utils/systemConsts';

import brandImage from 'utils/images/brand.svg';

import './app.scss';

const ROUTES_CONFIG = [
    {
		path: ROUTES.DEFAULT,
		component: () => "TBD",
        icon: ICON_NAMES.DASHBOARD,
        isIndex: true,
        title: "Dashboard"
	},
	{
		path: ROUTES.SCANS,
		component: Scans,
        icon: ICON_NAMES.SCANS,
        title: "Scans"
	},
	{
		path: ROUTES.ASSETS,
		component: () => "TBD",
        icon: ICON_NAMES.ASSETS,
        title: "Assets"
	},
	{
		path: ROUTES.FINDINGS,
		component: Findings,
        icon: ICON_NAMES.FINDINGS,
        title: "Findings"
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

const NavLinkItem = ({pathname, icon}) => {
    const location = useLocation();
    const match = useMatch(`${pathname}/*`);
    const isActive = pathname === location.pathname ? true : !!match;

    const navigate = useNavigate();

    const onClick = () => {
        navigate(pathname);
    }
    
    return (
        <div className={classnames("nav-item", {active: isActive})} onClick={onClick}>
            <Icon name={icon} />
        </div>
    )
}

const Layout = () => {
    const navigate = useNavigate()
    const {pathname} = useLocation();
    const mainPath = pathname.split("/").find(item => !!item);
    const pageTitle = ROUTES_CONFIG.find(({path, isIndex}) => (isIndex && !mainPath) || path === `/${mainPath}`)?.title;
    
    return (
        <div id="main-wrapper">  
            <div className="topbar-container">
                <img src={brandImage} alt="VMClarity" />
                {!!pageTitle &&
                    <div className="topbar-page-title">
                        <Title medium removeMargin>{pageTitle}</Title>
                        <Icon name={ICON_NAMES.REFRESH} onClick={() => navigate(0)} />
                    </div>
                }
            </div>
            <div className="sidebar-container">
                {
                    ROUTES_CONFIG.map(({path, icon, title}) => (
                        <TooltipWrapper key={path} tooltipId={`sidebar-item-tooltip-${path}`} tooltipText={title}>
                            <NavLinkItem pathname={path} icon={icon} />
                        </TooltipWrapper>
                    ))
                }
            </div>
            <main role="main">
                <NotificationProvider>
                    <Outlet />
                    <ConnectedNotification />
                </NotificationProvider>
            </main>
        </div>
    )
}

const App = () => (
    <div className="app-wrapper">
        
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<Layout />}>
                    {
                        ROUTES_CONFIG.map(({path, component: Component, isIndex}) => (
                            <Route key={path} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={(<Component />)} />
                        ))
                    }
                </Route>
            </Routes>
        </BrowserRouter>
        <IconTemplates />
    </div>
)

export default App;