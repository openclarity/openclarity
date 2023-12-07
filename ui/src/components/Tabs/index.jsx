import React from 'react';
import classnames from 'classnames';
import { TooltipWrapper } from 'components/Tooltip';

import './tabs.scss';

const Tabs = ({items, checkIsActive, onClick, headerCustomDisplay: HeaderCustomDisplay, withStickyTabs=false, tabItemPadding=20}) => (
    <div className={classnames("tabs-container", {sticky: withStickyTabs})}>
        {
            items.map((item) => {
                const {id, title, disabled, tabTooltip, customTitle: CustomTitle} = item;
                const isActive = checkIsActive(item);

                const onTabClick = () => {
                    if (disabled) {
                        return;
                    }
                    
                    onClick(item);
                }

                const WrapperElement = !!tabTooltip ? TooltipWrapper : "div";
                const wrapperProps = !!tabTooltip ? {tooltipId: `tab-disabled-tooltip-${id}`, tooltipText: tabTooltip} : {};
                
                return (
                    <WrapperElement key={id} {...wrapperProps} className={classnames("tab-item", {active: isActive}, {disabled})} style={{padding: `0 ${tabItemPadding}px`}} onClick={onTabClick}>
                        {!!CustomTitle ? <CustomTitle {...item} isActive={isActive} /> : title}
                    </WrapperElement>
                )
            })
        }
        {!!HeaderCustomDisplay && <div className="tabs-header-custom-content-wrapper"><HeaderCustomDisplay /></div>}
    </div>
);

export default Tabs;