import React from 'react';
import { Routes, Route, useNavigate, Outlet, useLocation, useParams } from 'react-router-dom';
import classnames from 'classnames';
import PageContainer from 'components/PageContainer';
import { TooltipWrapper } from 'components/Tooltip';

import './tabbed-page-container.scss';

const TabbedPageContainer = ({items}) => {
    const navigate = useNavigate();

    const {pathname} = useLocation();
    const params = useParams();

    const tabInnerPath = params["*"];
    const cleanPath = !!tabInnerPath ? pathname.replace(`/${tabInnerPath}`, "") : pathname;

    return (
        <PageContainer className="tabbed-page-container">
            <div className="tabs-container">
                {
                    items.map(({id, path, title, isIndex, disabled, tabTooltip}) => {
                        const isActive = (isIndex && pathname === cleanPath) || path === pathname.replace(`${cleanPath}/`, "");

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
            </div>
            <Routes>
                <Route path="/" element={<div className="tab-content"><Outlet /></div>}>
                    {
                        items.map(({id, path, isIndex, component: Component}) => (
                            <Route key={id} path={isIndex ? undefined : `${path}/*`} index={isIndex} element={<Component />} />
                        ))
                    }
                </Route>
            </Routes>
        </PageContainer>
    );
}

export default TabbedPageContainer;