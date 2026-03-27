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

  if (response.status === 204) {
    return null;
  }

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

export function updateProject(id, data) {
  return request('PUT', `/projects/${id}`, data);
}

export function createColumn(projectId, name) {
  return request('POST', `/projects/${projectId}/columns`, { name });
}

export function updateColumn(projectId, colId, name) {
  return request('PUT', `/projects/${projectId}/columns/${colId}`, { name });
}

export function deleteColumn(projectId, colId) {
  return request('DELETE', `/projects/${projectId}/columns/${colId}`);
}

export function reorderColumns(projectId, columnIds) {
  return request('PUT', `/projects/${projectId}/columns/reorder`, { columnIds });
}

export function createLabel(projectId, name, color) {
  return request('POST', `/projects/${projectId}/labels`, { name, color });
}

export function updateLabel(projectId, labelId, name, color) {
  return request('PUT', `/projects/${projectId}/labels/${labelId}`, { name, color });
}

export function deleteLabel(projectId, labelId) {
  return request('DELETE', `/projects/${projectId}/labels/${labelId}`);
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

export function updateTask(projectId, taskId, data) {
  return request('PUT', `/projects/${projectId}/tasks/${taskId}`, data);
}

export function moveTask(projectId, taskId, columnId, position) {
  return request('PUT', `/projects/${projectId}/tasks/${taskId}/move`, { columnId, position });
}

export function updateProfile(data) {
  return request('PUT', '/users/me', data);
}

export function adminListUsers() {
  return request('GET', '/admin/users');
}

export function adminCreateUser(data) {
  return request('POST', '/admin/users', data);
}

export function adminUpdateUser(userId, data) {
  return request('PUT', `/admin/users/${userId}`, data);
}

export function adminResetPassword(userId, password) {
  return request('PUT', `/admin/users/${userId}/password`, { password });
}

export function changePassword(data) {
  return request('PUT', '/users/me/password', data);
}

export function deleteTask(projectId, taskId) {
  return request('DELETE', `/projects/${projectId}/tasks/${taskId}`);
}

export function listComments(projectId, taskId) {
  return request('GET', `/projects/${projectId}/tasks/${taskId}/comments`);
}

export function createComment(projectId, taskId, text) {
  return request('POST', `/projects/${projectId}/tasks/${taskId}/comments`, { text });
}

export function updateComment(projectId, taskId, commentId, text) {
  return request('PUT', `/projects/${projectId}/tasks/${taskId}/comments/${commentId}`, { text });
}

export function deleteComment(projectId, taskId, commentId) {
  return request('DELETE', `/projects/${projectId}/tasks/${taskId}/comments/${commentId}`);
}
