import React from 'react';
import classnames from 'classnames';
import { isEmpty } from 'lodash';
import { ICON_NAMES } from './utils';
import IconTemplates from './IconTemplates';

import './icon.scss';

export {
	ICON_NAMES,
	IconTemplates
}

const Icon = ({name, className, onClick, disabled, style={}}) => {
	if (!Object.values(ICON_NAMES).includes(name)) {
		console.error(`Icon name '${name}' does not exist`);
	}
	
	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			xmlnsXlink="http://www.w3.org/1999/xlink"
			className={classnames(
				"icon",
				`icon-${name}`,
				{[className]: !isEmpty(className)},
				{clickable: !!onClick && !disabled},
				{disabled}
			)}
			onClick={event => !disabled && !!onClick ? onClick(event) : undefined}
			style={style}
		>
			<use href={`#${name}`} />
		</svg>
	)
}

export default Icon;