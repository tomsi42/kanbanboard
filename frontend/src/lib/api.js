const BASE = '/api/v1';

async function request(method, path, body = null) {
  const options = {
    method,
    headers: {},
  };

  if (body) {
    options.headers['Content-Type'] = 'application/json';
    options.body = JSON.stringify(body);
  }

  const response = await fetch(`${BASE}${path}`, options);
  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || 'Request failed');
  }

  return data;
}

export function getSetupStatus() {
  return request('GET', '/setup/status');
}

export function postSetup(data) {
  return request('POST', '/setup', data);
}

export function getAppTitle() {
  return request('GET', '/app/title');
}

export function getHealth() {
  return request('GET', '/health');
}

export function login(email, password) {
  return request('POST', '/auth/login', { email, password });
}

export function logout() {
  return request('POST', '/auth/logout');
}

export function getMe() {
  return request('GET', '/auth/me');
}

export function listProjects() {
  return request('GET', '/projects');
}

export function getProject(id) {
  return request('GET', `/projects/${id}`);
}

export function createProject(name) {
  return request('POST', '/projects', { name });
}

export function listTasks(projectId) {
  return request('GET', `/projects/${projectId}/tasks`);
}

export function createTask(projectId, data) {
  return request('POST', `/projects/${projectId}/tasks`, data);
}
