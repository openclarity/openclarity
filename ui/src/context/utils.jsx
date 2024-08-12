import React from 'react';
import { isUndefined } from 'lodash';

export const create = (reducer, initialState) => {
    const StateContext = React.createContext();
    const DispatchContext = React.createContext();

    const Provider = ({children}) => {
        const [state, dispatch] = React.useReducer(reducer, initialState);

        return (
            <StateContext.Provider value={state}>
                <DispatchContext.Provider value={dispatch}>
                    {children}
                </DispatchContext.Provider>
            </StateContext.Provider>
        )
    }

    const useState = () => {
        const context = React.useContext(StateContext);

        if (isUndefined(context)) {
            throw Error("useState is not within the related provider")
        }

        return context;
    }

    const useDispatch = () => {
        const context = React.useContext(DispatchContext);

        if (isUndefined(context)) {
            throw Error("useDispatch is not within the related provider")
        }

        return context;
    }

    return [Provider, useState, useDispatch];
}