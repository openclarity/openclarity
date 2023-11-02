import React, { useEffect, useRef, useState } from "react";
import Arrow, { ARROW_NAMES } from 'components/Arrow';
import { TooltipWrapper } from "components/Tooltip";

import './wrapping-text-box-with-ellipsis.scss';

export const WrappingTextBoxWithEllipsis = ({ children, numberOfLines = 1 }) => {
    const [isExpanded, setIsExpanded] = useState(false);
    const [showExpandButton, setShowExpandButton] = useState(false);
    const ref = useRef();

    useEffect(() => {
        setShowExpandButton(ref.current && (ref.current.offsetHeight < ref.current.scrollHeight || ref.current.offsetWidth < ref.current.scrollWidth));
    }, [children]);

    return (
        <div className="wrapping-text-box-with-ellipsis-wrapper">
            <div ref={ref} className="wrapping-text-box-with-ellipsis-content" style={{
                WebkitLineClamp: isExpanded ? 'initial' : `${numberOfLines}`,
            }}>
                {children}
            </div>
            {showExpandButton &&
                <div>
                    <TooltipWrapper tooltipId={children} tooltipText={isExpanded ? 'Show less' : 'Show more'}>
                        <Arrow
                            name={isExpanded ? ARROW_NAMES.TOP : ARROW_NAMES.BOTTOM}
                            onClick={() => setIsExpanded(!isExpanded)}
                            small
                        />
                    </TooltipWrapper>
                </div>
            }
        </div>
    )
}
