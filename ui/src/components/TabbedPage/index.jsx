import React from 'react';
import { Routes, Route, useNavigate, Outlet, useLocation, useParams, Navigate } from 'react-router-dom';
import classnames from 'classnames';
import Tabs from 'components/Tabs';

import './tabbed-page.scss';

const TabbedPage= ({items, redirectTo, basePath, headerCustomDisplay, withInnerPadding=true, withStickyTabs=false}) => {
    const navigate = useNavigate();

    const {pathname} = useLocation();
    const params = useParams();

    const tabInnerPath = params["*"];
    const cleanPath = !!basePath ? basePath : (!!tabInnerPath ? pathname.replace(`/${tabInnerPath}`, "") : pathname);

    const checkIsActive = ({isIndex, path}) => (isIndex && pathname === cleanPath) || path === pathname.replace(`${cleanPath}/`, "");

    const isInTab = items.find(({path, isIndex}) => checkIsActive({isIndex, path}));
    
    return (
        <div className="tabbed-page-container">
            {isInTab &&
                <Tabs
                    items={items}
                    checkIsActive={checkIsActive}
                    onClick={({isIndex, path}) => navigate(isIndex ? cleanPath : path)}
                    headerCustomDisplay={headerCustomDisplay}
                    withStickyTabs={withStickyTabs}
                />
            }
            <Routes>
                <Route path="/" element={<div className={classnames("tab-content", {"with-padding": withInnerPadding})}><Outlet /></div>}>
                    {
                        items.map(({id, path, isIndex, component: Component}) => (
                            <Route key={id} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={<Component />} />
                        ))
                    }
                    {!!redirectTo && <Route path="" element={<Navigate to={redirectTo} replace />} />}
                </Route>
            </Routes>
        </div>
    );
}

export default TabbedPage;
