import * as React from "react";
import Axios from "axios";

interface CommitInfo {
  hash: string;
  message: string;
  author: string;
  date: string;
}

interface StatusResponse {
  ahead: number;
  commits: CommitInfo[];
  remote: string;
}

interface ActionResponse {
  success: boolean;
  message: string;
}

interface GitState {
  loading: boolean;
  pushing: boolean;
  pulling: boolean;
  error: string | null;
  actionResult: string | null;
  status: StatusResponse | null;
}

export default class GitPage extends React.Component<{}, GitState> {
  constructor(props) {
    super(props);
    this.state = {
      loading: true,
      pushing: false,
      pulling: false,
      error: null,
      actionResult: null,
      status: null
    };
  }

  componentDidMount() {
    this.loadStatus();
  }

  private loadStatus() {
    this.setState({loading: true, error: null});
    Axios.get('/api/v1/git/' + this.props.entity + '/status')
      .then(res => {
        this.setState({loading: false, status: res.data});
      })
      .catch(err => {
        this.setState({loading: false, error: 'Failed to load git status'});
        handleErr(err);
      });
  }

  private doPush() {
    this.setState({pushing: true, actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/push')
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({pushing: false, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({pushing: false, error: 'Push failed'});
        handleErr(err);
      });
  }

  private doPull() {
    this.setState({pulling: true, actionResult: null, error: null});
    Axios.post('/api/v1/git/' + this.props.entity + '/pull')
      .then(res => {
        const result: ActionResponse = res.data;
        this.setState({pulling: false, actionResult: result.message});
        if (result.success) {
          this.loadStatus();
        }
      })
      .catch(err => {
        this.setState({pulling: false, error: 'Pull failed'});
        handleErr(err);
      });
  }

  render() {
    let content = null;

    if (this.state.loading) {
      content = <p><i className="fa fa-spinner fa-spin"></i> Loading git status...</p>;
    } else if (this.state.error) {
      content = <div className="alert alert-danger">{this.state.error}</div>;
    } else if (this.state.status) {
      const status = this.state.status;

      let commitsList = null;
      if (status.commits.length > 0) {
        commitsList = <table className="table table-striped">
          <thead>
            <tr>
              <th>Hash</th>
              <th>Message</th>
              <th>Author</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {status.commits.map((c, i) => (
              <tr key={i}>
                <td><code>{c.hash}</code></td>
                <td>{c.message}</td>
                <td>{c.author}</td>
                <td>{c.date}</td>
              </tr>
            ))}
          </tbody>
        </table>;
      } else {
        commitsList = <p className="text-muted">No unpushed commits. Everything is up to date.</p>;
      }

      content = <div>
        <div className="row m-b-md">
          <div className="col-md-6">
            <p><strong>Remote:</strong> {status.remote || 'Not configured'}</p>
            <p><strong>Commits ahead:</strong> {status.ahead}</p>
          </div>
          <div className="col-md-6 text-right">
            <button
              className="btn btn-default m-r-sm"
              onClick={this.doPull.bind(this)}
              disabled={this.state.pulling || this.state.pushing}
            >
              {this.state.pulling ? (
                <span><i className="fa fa-spinner fa-spin"></i> Pulling...</span>
              ) : (
                <span><i className="fa fa-cloud-download"></i> Pull from Remote</span>
              )}
            </button>
            <button
              className="btn btn-primary"
              onClick={this.doPush.bind(this)}
              disabled={this.state.pushing || this.state.pulling || status.ahead === 0}
            >
              {this.state.pushing ? (
                <span><i className="fa fa-spinner fa-spin"></i> Pushing...</span>
              ) : (
                <span><i className="fa fa-cloud-upload"></i> Push to Remote</span>
              )}
            </button>
          </div>
        </div>

        {this.state.actionResult && (
          <div className="alert alert-info">{this.state.actionResult}</div>
        )}

        <h4>Unpushed Commits</h4>
        {commitsList}
      </div>;
    }

    return <div className="normalheader">
      <div className="hpanel hblue">
        <div className="panel-heading hbuilt">
          <i className="fa fa-git"></i> Git Status
          <div className="panel-tools">
            <button className="btn btn-default btn-xs" onClick={this.loadStatus.bind(this)}>
              <i className="fa fa-refresh"></i> Refresh
            </button>
          </div>
        </div>
        <div className="panel-body">
          {content}
        </div>
      </div>
    </div>;
  }
}
