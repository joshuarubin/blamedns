import React, { PropTypes, Component } from 'react'
import { connect } from 'react-redux'
import Log from './Log'

class Logs extends Component {
    render() {
        let entries = []
        for (let i = this.props.entries.length-1; i >=0; i--) {
            entries.push(this.props.entries[i])
        }

        return (
            <div>
                {entries.map((entry, idx) => <Log key={idx} entry={entry} />)}
            </div>
        )
    }
}

Logs.propTypes = {
    entries: PropTypes.arrayOf(PropTypes.shape({
        Data: PropTypes.object,
        Level: PropTypes.number.isRequired,
        Time: PropTypes.string.isRequired,
        Message: PropTypes.string.isRequired
    }).isRequired).isRequired,
    logLevel: PropTypes.number.isRequired,
}

const getVisibleLogEntries = (logs, logLevel) => {
    return logs.filter(l => l.Level <= logLevel)
}

const mapStateToProps = (state, ownProps) => {
    return {
        entries: getVisibleLogEntries(state.logs, ownProps.logLevel)
    }
}

export default connect(mapStateToProps)(Logs)
