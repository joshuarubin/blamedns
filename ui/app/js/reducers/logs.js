const logs = (state = [], action) => {
    switch (action.type) {
    case 'ADD_LOG_ENTRY':
        return [
            action.data,
            ...state.slice(0, 1023)
        ]
    default:
        return state
    }
}

export default logs
