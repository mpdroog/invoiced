import * as React from "react";
import Axios from "axios";
import "./search.css";
import { ActionButton } from "./ActionButton";
import { ModalBackdrop, openModal, closeModal } from "./Modal";
import type { StatusResponse } from "../types/git";

interface SearchResponse {
    results: SearchResult[];
}

interface SearchResult {
    type: string;
    id: string;
    title: string;
    subtitle: string;
    entity: string;
    year: number;
    quarter: number;
    bucket: string;
}

interface MenuState {
    ahead: number;
    searchQuery: string;
    searchResults: SearchResult[];
    showResults: boolean;
    searching: boolean;
    menuOpen: boolean;
}

interface MenuProps {
    company: string;
    year: string;
}

class Menu extends React.Component<MenuProps, MenuState> {
    private boundRefresh: () => void;
    private searchTimeout: ReturnType<typeof setTimeout> | null = null;

    constructor(props: MenuProps) {
        super(props);
        this.state = {
            ahead: 0,
            searchQuery: '',
            searchResults: [],
            showResults: false,
            searching: false,
            menuOpen: false
        };
        this.boundRefresh = this.fetchGitStatus.bind(this);
    }

    componentDidMount(): void {
        this.fetchGitStatus();
        window.addEventListener('git-refresh', this.boundRefresh);
        document.addEventListener('click', this.handleClickOutside.bind(this));
    }

    componentWillUnmount(): void {
        window.removeEventListener('git-refresh', this.boundRefresh);
        document.removeEventListener('click', this.handleClickOutside.bind(this));
    }

    private handleClickOutside(e: MouseEvent): void {
        const searchContainer = document.getElementById('search-container');
        const target = e.target;
        if (searchContainer && target instanceof Node && !searchContainer.contains(target)) {
            this.setState({showResults: false});
        }
    }

    private fetchGitStatus(): void {
        Axios.get<StatusResponse>('/api/v1/git/' + this.props.company + '/status')
            .then(res => {
                this.setState({ahead: res.data.ahead});
            })
            .catch(() => {
                // Silently ignore - git status may not be available
            });
    }

    private handleSearchChange(e: React.ChangeEvent<HTMLInputElement>): void {
        const query = e.target.value;
        this.setState({searchQuery: query, showResults: true});

        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }

        if (query.length < 2) {
            this.setState({searchResults: [], searching: false});
            return;
        }

        this.setState({searching: true});
        this.searchTimeout = setTimeout(() => {
            Axios.get<SearchResponse>('/api/v1/search/' + this.props.company, {params: {q: query}})
                .then(res => {
                    this.setState({searchResults: res.data.results, searching: false});
                })
                .catch(() => {
                    this.setState({searchResults: [], searching: false});
                });
        }, 300);
    }

    private handleResultClick(result: SearchResult): void {
        let url = '#' + result.entity + '/' + result.year + '/';
        if (result.type === 'invoice') {
            url += 'invoices/edit/' + result.bucket + '/' + result.id;
        } else if (result.type === 'hour') {
            url += 'hours/edit/' + result.bucket + '/' + result.id;
        } else if (result.type === 'purchase') {
            url += 'purchases';
        }
        this.setState({showResults: false, searchQuery: ''});
        location.hash = url;
    }

    private getResultIcon(type: string): string {
        switch (type) {
            case 'invoice': return 'fa-money-bill';
            case 'hour': return 'fa-clock';
            case 'purchase': return 'fa-shopping-cart';
            default: return 'fa-file';
        }
    }

    private async handleLogout(): Promise<void> {
        await Axios.post('/logout');
        window.location.href = '/';
    }

    render(): React.JSX.Element {
        const {company, year} = this.props;
        const {ahead, searchQuery, searchResults, showResults, searching, menuOpen} = this.state;

        return (
            <nav className="navbar navbar-expand-md bg-body-tertiary fixed-top">
                <div className="container">
                    <a href="/" className="navbar-brand" id="js-entities">
                        <i className="fas fa-calculator me-2"></i>
                        <span className="fw-bold">Boekhoud</span><span className="text-success">.cloud</span>
                    </a>

                    <button
                        className="navbar-toggler"
                        type="button"
                        aria-label="Toggle navigation"
                        onClick={() => this.setState({menuOpen: !menuOpen})}
                    >
                        <span className="navbar-toggler-icon"></span>
                    </button>

                    <div className={"collapse navbar-collapse" + (menuOpen ? " show" : "")}>
                        <div id="search-container" className="me-auto search-container">
                            <div className="form-group">
                                <input
                                    type="text"
                                    placeholder="Search invoices, hours..."
                                    className="form-control"
                                    id="js-search"
                                    value={searchQuery}
                                    onChange={this.handleSearchChange.bind(this)}
                                    onFocus={() => this.setState({showResults: true})}
                                />
                            </div>
                            {showResults && searchQuery.length >= 2 && (
                                <div className="search-dropdown">
                                    {searching ? (
                                        <div className="search-loading">
                                            <i className="fas fa-spinner fa-spin"></i> Searching...
                                        </div>
                                    ) : searchResults.length > 0 ? (
                                        searchResults.map((r, i) => (
                                            <div
                                                key={i}
                                                className="search-result-item"
                                                onClick={() => this.handleResultClick(r)}
                                            >
                                                <i className={'fas ' + this.getResultIcon(r.type) + ' search-result-icon'}></i>
                                                <strong>{r.title}</strong>
                                                <span className="search-result-subtitle">{r.subtitle}</span>
                                                <div className="search-result-meta">
                                                    {r.type} &middot; {r.year} Q{r.quarter}
                                                </div>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="search-no-results">
                                            No results found
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>

                        <ul className="navbar-nav ms-auto">
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year} className="nav-link" id="js-dashboard" title="Dashboard">
                                    <i className="fas fa-gauge fa-lg"></i>
                                    <span className="d-md-none ms-2">Dashboard</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year+"/hours"} className="nav-link" id="js-hours" title="Hours">
                                    <i className="far fa-clock fa-lg"></i>
                                    <span className="d-md-none ms-2">Hours</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year+"/invoices"} className="nav-link" id="js-invoices" title="Invoices">
                                    <i className="fas fa-money-bill fa-lg"></i>
                                    <span className="d-md-none ms-2">Invoices</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year+"/purchases"} className="nav-link" id="js-purchases" title="Purchases">
                                    <i className="fas fa-shopping-cart fa-lg"></i>
                                    <span className="d-md-none ms-2">Purchases</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year+"/taxes"} className="nav-link" id="js-taxes" title="Taxes">
                                    <i className="fas fa-building-columns fa-lg"></i>
                                    <span className="d-md-none ms-2">Taxes</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a href={"#"+company+"/"+year+"/git"} className="nav-link" id="js-git" title="Git">
                                    <i className="fas fa-cloud-arrow-up fa-lg"></i>
                                    <span className="d-md-none ms-2">Sync</span>
                                    {ahead > 0 && <span className="badge bg-danger">{ahead}</span>}
                                </a>
                            </li>
                            <li className="nav-item">
                                <ActionButton
                                    className="nav-link btn btn-link"
                                    onClick={() => this.handleLogout()}
                                    title="Logout"
                                >
                                    <i className="fas fa-sign-out-alt fa-lg"></i>
                                    <span className="d-md-none ms-2">Logout</span>
                                </ActionButton>
                            </li>
                        </ul>
                    </div>
                </div>
            </nav>
        );
    }
}

export function calendar({ year }: { year: string }): React.JSX.Element {
    return (<div>
        <div className="month">
          <ul>
            <li className="prev">&#10094;</li>
            <li className="next">&#10095;</li>
            <li>
              August<br/>
              <span className="calendar-year">{year}</span>
            </li>
          </ul>
        </div>

        <ul className="weekdays">
          <li>Mo</li>
          <li>Tu</li>
          <li>We</li>
          <li>Th</li>
          <li>Fr</li>
          <li>Sa</li>
          <li>Su</li>
        </ul>

        <ul className="days"> 
          <li>1</li>
          <li>2</li>
          <li>3</li>
          <li>4</li>
          <li>5</li>
          <li>6</li>
          <li>7</li>
          <li>8</li>
          <li>9</li>
          <li><span className="active">10</span></li>
          <li>11</li>
        </ul>
    </div>);
}

interface LoginModalState {
    visible: boolean;
    email: string;
    password: string;
    error: string;
}

class LoginModal extends React.Component<Record<string, never>, LoginModalState> {
    private boundShow: () => void;

    constructor(props: Record<string, never>) {
        super(props);
        this.state = {
            visible: false,
            email: '',
            password: '',
            error: ''
        };
        this.boundShow = (): void => {
            openModal();
            this.setState({visible: true, error: '', password: ''});
        };
    }

    componentDidMount(): void {
        window.addEventListener('show-login-modal', this.boundShow);
    }

    componentWillUnmount(): void {
        window.removeEventListener('show-login-modal', this.boundShow);
    }

    private async handleLogin(): Promise<void> {
        this.setState({error: ''});

        const params = new URLSearchParams();
        params.append('email', this.state.email);
        params.append('pass', this.state.password);

        const res = await Axios.post('/', params, {
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            validateStatus: (status) => status < 500
        });

        if (res.status === 200) {
            closeModal();
            window.dispatchEvent(new Event('login-modal-closed'));
            window.location.reload();
        } else {
            this.setState({error: 'Invalid credentials'});
        }
    }

    render(): React.JSX.Element | null {
        if (!this.state.visible) return null;

        return (
            <div className="modal in modal-show" role="dialog">
                    <div className="modal-dialog">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h4 className="modal-title">
                                    <i className="fas fa-lock"></i> Session Expired
                                </h4>
                            </div>
                            <div className="modal-body">
                                <p>Your session has expired. Please log in again to continue.</p>
                                {this.state.error !== '' && (
                                    <div className="alert alert-danger">{this.state.error}</div>
                                )}
                                <div className="form-group">
                                    <input
                                        type="email"
                                        className="form-control"
                                        placeholder="Email"
                                        value={this.state.email}
                                        onChange={e => this.setState({email: e.target.value})}
                                        autoFocus
                                    />
                                </div>
                                <div className="form-group">
                                    <input
                                        type="password"
                                        className="form-control"
                                        placeholder="Password"
                                        value={this.state.password}
                                        onChange={e => this.setState({password: e.target.value})}
                                    />
                                </div>
                            </div>
                            <div className="modal-footer">
                                <ActionButton
                                    className="btn btn-success"
                                    onClick={() => this.handleLogin()}
                                >
                                    Login
                                </ActionButton>
                            </div>
                        </div>
                    </div>
                </div>
        );
    }
}

interface DesignProps {
    entity: string;
    year: string;
    children?: React.ReactNode;
}

export class Design extends React.Component<DesignProps> {
  constructor(props: DesignProps) {
    super(props);
  }
  render(): React.JSX.Element {
    return <div>
        <Menu company={this.props.entity} year={this.props.year}/>
        <div className="container-fluid">
            {this.props.children}
        </div>
        <LoginModal />
        <ModalBackdrop />
    </div>;
  }
}
