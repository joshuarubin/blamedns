import React, { PropTypes, Component } from 'react'
import { connect } from 'react-redux'
import Log from './Log'

class Logs extends Component {
    render() {
        return (
            <table className="table table-hover log-messages table-condensed">
                <thead><tr>
                    <th>Level</th>
                    <th>Time</th>
                    <th>Message</th>
                    <th>Data</th>
                </tr></thead>
                <tbody>
                    {this.props.entries.map((entry, idx) => <Log key={idx} entry={entry} />)}
                </tbody>
            </table>
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
