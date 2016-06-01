const logs = (state = [], action) => {
    switch (action.type) {
    case 'ADD_LOG_ENTRY':
        return [
            ...state,
            action.data
        ]
    default:
        return state
    }
}

export default logs
