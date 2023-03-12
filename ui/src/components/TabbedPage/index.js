import React from 'react';
import { Routes, Route, useNavigate, Outlet, useLocation, useParams, Navigate } from 'react-router-dom';
import classnames from 'classnames';
import { TooltipWrapper } from 'components/Tooltip';

import './tabbed-page.scss';

const TabbedPage= ({items, redirectTo, basePath, headerCustomDisplay: HeaderCustomDisplay, withInnerPadding=true, withStickyTabs=false}) => {
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
                <div className={classnames("tabs-container", {sticky: withStickyTabs})}>
                    {
                        items.map(({id, path, title, isIndex, disabled, tabTooltip}) => {
                            const isActive = checkIsActive({isIndex, path});

                            const onTabClick = () => {
                                if (disabled) {
                                    return;
                                }
                                
                                navigate(isIndex ? cleanPath : path);
                            }

                            const WrapperElement = !!tabTooltip ? TooltipWrapper : "div";
                            const wrapperProps = !!tabTooltip ? {tooltipId: `tab-disabled-tooltip-${id}`, tooltipText: tabTooltip} : {};
                            
                            return (
                                <WrapperElement key={id} {...wrapperProps} className={classnames("tab-item", {active: isActive}, {disabled})} onClick={onTabClick}>
                                    {title}
                                </WrapperElement>
                            )
                        })
                    }
                    {!!HeaderCustomDisplay && <div className="tabs-header-custom-content-wrapper"><HeaderCustomDisplay /></div>}
                </div>
            }
            <Routes>
                <Route path="/" element={<div className={classnames("tab-content", {"with-padding": withInnerPadding})}><Outlet /></div>}>
                    {
                        items.map(({id, path, isIndex, component: Component}) => (
                            <Route key={id} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={<Component />} />
                        ))
                    }
                    {!!redirectTo && <Route path="" element={<Navigate to={redirectTo} />} />}
                </Route>
            </Routes>
        </div>
    );
}

export default TabbedPage;