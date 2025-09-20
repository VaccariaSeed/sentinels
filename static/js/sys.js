// 全局变量
let currentPage = {
    devices: 1,
    points: 1
};
const pageSize = 20;
let totalPoints = 0;
let currentDeviceId = '';

// 初始化函数
document.addEventListener('DOMContentLoaded', function () {
    loadDevices();
    setupEventListeners();
});

// 设置事件监听器
function setupEventListeners() {
    // 标签切换功能
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));

            tab.classList.add('active');
            const tabId = tab.getAttribute('data-tab');
            document.getElementById(tabId).classList.add('active');

            // 加载对应标签的数据
            if (tabId === 'device-config') {
                loadDevices();
            } else if (tabId === 'point-config') {
                loadPoints();
            } else if (tabId === 'system-monitor') {
                fetchDeviceData();
            }
        });
    });

    // 模态框功能
    const deviceModal = document.getElementById('device-modal');
    const pointModal = document.getElementById('point-modal');
    const collectionRuleModal = document.getElementById('collection-rule-modal');
    const ruleEditModal = document.getElementById('rule-edit-modal');
    const addDeviceBtn = document.getElementById('add-device-btn');
    const addPointBtn = document.getElementById('add-point-btn');
    const collectionRuleBtn = document.getElementById('collection-rule-btn');
    const addRuleBtn = document.getElementById('add-rule-btn');
    const closeButtons = document.querySelectorAll('.close');
    const cancelDeviceBtn = document.getElementById('cancel-device');
    const cancelPointBtn = document.getElementById('cancel-point');
    const cancelRuleBtn = document.getElementById('cancel-rule');

    // 打开设备模态框
    addDeviceBtn.addEventListener('click', () => {
        document.getElementById('device-form').reset();
        document.getElementById('device-id').value = '';
        deviceModal.style.display = 'flex';
    });

    // 打开点位模态框
    addPointBtn.addEventListener('click', () => {
        document.getElementById('point-form').reset();
        document.getElementById('point-id').value = '';
        pointModal.style.display = 'flex';
    });

    // 打开采集规则模态框
    collectionRuleBtn.addEventListener('click', () => {
        loadCollectionRules();
        collectionRuleModal.style.display = 'flex';
    });

    // 打开添加规则模态框
    addRuleBtn.addEventListener('click', () => {
        document.getElementById('rule-edit-form').reset();
        document.getElementById('rule-id').value = '';
        document.getElementById('rule-edit-title').textContent = '添加采集规则';
        ruleEditModal.style.display = 'flex';
    });

    // 关闭模态框
    function closeModals() {
        deviceModal.style.display = 'none';
        pointModal.style.display = 'none';
        collectionRuleModal.style.display = 'none';
        ruleEditModal.style.display = 'none';
    }

    // 点击关闭按钮
    closeButtons.forEach(btn => {
        btn.addEventListener('click', closeModals);
    });

    // 点击取消按钮
    cancelDeviceBtn.addEventListener('click', closeModals);
    cancelPointBtn.addEventListener('click', closeModals);
    cancelRuleBtn.addEventListener('click', closeModals);

    // 点击模态框外部关闭
    window.addEventListener('click', (e) => {
        if (e.target === deviceModal) closeModals();
        if (e.target === pointModal) closeModals();
        if (e.target === collectionRuleModal) closeModals();
        if (e.target === ruleEditModal) closeModals();
    });

    // 表单提交
    document.getElementById('device-form').addEventListener('submit', (e) => {
        e.preventDefault();
        saveDevice();
    });

    document.getElementById('point-form').addEventListener('submit', (e) => {
        e.preventDefault();
        savePoint();
    });

    document.getElementById('rule-edit-form').addEventListener('submit', (e) => {
        e.preventDefault();
        saveCollectionRule();
    });

    document.getElementById('search-btn').addEventListener('click', function () {
        currentPage.points = 1;
        loadPoints();
    });

    // 用户操作按钮
    document.querySelectorAll('.user-action-card button').forEach(button => {
        button.addEventListener('click', function () {
            const action = this.closest('.user-action-card').querySelector('h3').textContent;
            executeUserAction(action);
        });
    });

    // 系统监控相关事件
    setupSystemMonitorEvents();
}

// 设置系统监控事件
function setupSystemMonitorEvents() {
    // 获取DOM元素
    const refreshBtn = document.getElementById('refreshBtn');
    const closeBtn = document.querySelector('#alarmModal .close-btn');

    // 刷新数据
    refreshBtn.addEventListener('click', function () {
        fetchDeviceData();
    });

    // 关闭告警模态框
    closeBtn.addEventListener('click', function () {
        document.getElementById('alarmModal').style.display = 'none';
    });

    // 点击模态框外部关闭
    window.addEventListener('click', function (event) {
        const alarmModal = document.getElementById('alarmModal');
        if (event.target === alarmModal) {
            alarmModal.style.display = 'none';
        }
    });
}

// 加载设备列表
async function loadDevices() {
    try {
        showLoading('device-config');

        // 调用API获取设备列表
        const response = await fetch('/api/devices', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const devices = await response.json();
        renderDevicesTable(devices);

        // 更新设备下拉选择框
        updateDeviceSelect(devices);
    } catch (error) {
        showNotification('加载设备列表失败' + error, 'error');
    }
}

// 更新设备下拉选择框
function updateDeviceSelect(devices) {
    const deviceSelect = document.getElementById('device-select');
    deviceSelect.innerHTML = '<option value="">请选择</option>';

    devices.forEach(device => {
        const option = document.createElement('option');
        option.value = device.id;
        option.textContent = `${device.name}`;
        deviceSelect.appendChild(option);
    });

    // 如果有设备ID，则选中对应设备
    if (currentDeviceId) {
        deviceSelect.value = currentDeviceId;
    }
}

// 渲染设备表格
function renderDevicesTable(devices) {
    const tbody = document.querySelector('#device-config tbody');
    tbody.innerHTML = '';

    if (devices.length === 0) {
        tbody.innerHTML = '<tr><td colspan="16" style="text-align: center;">暂无设备数据</td></tr>';
        return;
    }

    devices.forEach(device => {
        const row = document.createElement('tr');

        row.innerHTML = `
                    <td>
                        <label class="toggle-switch">
                            <input type="checkbox" ${device.status ? 'checked' : ''}
                                onchange="toggleDeviceStatus('${device.id}', this.checked)">
                            <span class="slider"></span>
                        </label>
                        <span class="device-status ${device.status ? 'status-active' : 'status-inactive'}">
                            ${device.status ? '运行中' : '已停止'}
                        </span>
                    </td>
                    <td>${device.id}</td>
                    <td>${device.name}</td>
                    <td>${device.code}</td>
                    <td>${device.table}</td>
                    <td>${device.interfaceType}</td>
                    <td>${device.address}</td>
                    <td>${device.baudRate}</td>
                    <td>${device.stopBits}</td>
                    <td>${device.parity === 'E' ? '奇校验' : device.parity === 'O' ? '偶校验' : '无校验'}</td>
                    <td>${device.dataBits}</td>
                    <td>${device.protocolType}</td>
                    <td>${device.deviceAddress}</td>
                    <td>${device.writeTimeout}</td>
                    <td>${device.readTimeout}</td>
                    <td>
                        <button class="btn btn-primary btn-sm" onclick="editDevice('${device.id}')">
                            修改
                        </button>
                        <button class="btn btn-danger btn-sm" onclick="deleteDevice('${device.id}')">
                            删除
                        </button>
                    </td>
                `;

        tbody.appendChild(row);
    });
}

// 加载点位列表（带分页）
async function loadPoints() {
    try {
        showLoading('point-config');

        const deviceId = document.getElementById('device-select').value;
        const deviceMark = document.getElementById('device-mark').value;
        let url = `/api/points?page=${currentPage.points}&pageSize=${pageSize}`;

        if (deviceId) {
            url += `&deviceId=${deviceId}`;
        }
        if (deviceMark) {
            url += `&deviceMark=${deviceMark}`;
        }

        // 调用API获取点位列表
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        renderPointsTable(data.points);

        // 更新分页信息
        totalPoints = data.totalCount;
        updatePointsPagination();
    } catch (error) {
        showNotification('加载点位列表失败'+ error, 'error');
    }
}

// 渲染点位表格
function renderPointsTable(points) {
    const tbody = document.querySelector('#point-config tbody');
    tbody.innerHTML = '';

    if (points.length === 0) {
        tbody.innerHTML = '<tr><td colspan="20" style="text-align: center;">暂无点位数据</td></tr>';
        return;
    }

    points.forEach(point => {
        const row = document.createElement('tr');

        row.innerHTML = `
                    <td>${point.id}</td>
                    <td>${point.functionCode}</td>
                    <td>${point.address}</td>
                    <td>${point.dataType}</td>
                    <td>${point.tag}</td>
                    <td>${point.luaExpression || ''}</td>
                    <td>${point.description}</td>
                    <td>${point.alarmFlag || ''}</td>
                    <td>${point.alarmLevel === 'serious' ? '严重' :
            point.alarmLevel === 'high' ? '高' :
                point.alarmLevel === 'middle' ? '中' :
                    point.alarmLevel === 'low' ? '低' : '不关注'}</td>
                    <td>${point.multiplier || 1}</td>
                    <td>${point.unit || ''}</td>
                    <td>${point.priority === 3 ? '高' :
            point.priority === 2 ? '中' :
                point.priority === 1 ? '低' : '未知'}</td>
                    <td>${point.endianness === 'LITTLE' ? '小端' : '大端'}</td>
                    <td>${point.bitCalculation === 'single' ? '单bit位' : point.bitCalculation === 'multiple' ? '多bit位' : '整算'}</td>
                    <td>${point.startBit || ''}</td>
                    <td>${point.endBit || ''}</td>
                    <td>${point.storageMethod === 'direct' ? '直存' : '变存'}</td>
                    <td>${point.offset}</td>
                    <td>${point.store || ''}</td>
                    <td>
                        <button class="btn btn-primary btn-sm" onclick="editPoint('${point.id}')">
                            修改
                        </button>
                        <button class="btn btn-danger btn-sm" onclick="deletePoint('${point.id}')">
                            删除
                        </button>
                    </td>
                `;

        tbody.appendChild(row);
    });
}

// 更新点位分页控件
function updatePointsPagination() {
    const paginationContainer = document.querySelector('#point-config .pagination');
    paginationContainer.innerHTML = '';

    const totalPages = Math.ceil(totalPoints / pageSize);

    // 添加上一页按钮
    if (currentPage.points > 1) {
        const prevButton = document.createElement('button');
        prevButton.textContent = '上一页';
        prevButton.addEventListener('click', () => {
            currentPage.points--;
            loadPoints();
        });
        paginationContainer.appendChild(prevButton);
    }

    // 添加页码按钮
    const startPage = Math.max(1, currentPage.points - 2);
    const endPage = Math.min(totalPages, startPage + 4);

    for (let i = startPage; i <= endPage; i++) {
        const pageButton = document.createElement('button');
        pageButton.textContent = i;
        pageButton.classList.toggle('active', i === currentPage.points);
        pageButton.addEventListener('click', () => {
            currentPage.points = i;
            loadPoints();
        });
        paginationContainer.appendChild(pageButton);
    }

    // 添加下一页按钮
    if (currentPage.points < totalPages) {
        const nextButton = document.createElement('button');
        nextButton.textContent = '下一页';
        nextButton.addEventListener('click', () => {
            currentPage.points++;
            loadPoints();
        });
        paginationContainer.appendChild(nextButton);
    }

    // 添加总页数信息
    const pageInfo = document.createElement('span');
    pageInfo.textContent = `共 ${totalPoints} 条记录，${totalPages} 页`;
    pageInfo.style.marginLeft = '15px';
    paginationContainer.appendChild(pageInfo);
}

// 加载采集规则
async function loadCollectionRules() {
    try {
        const deviceId = document.getElementById('device-select').value;
        const response = await fetch('/api/collection-rules' + `?deviceId=${deviceId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const rules = await response.json();
        renderCollectionRulesTable(rules);
    } catch (error) {
        showNotification('加载采集规则失败' + error, 'error');
    }
}

// 渲染采集规则表格
function renderCollectionRulesTable(rules) {
    const tbody = document.getElementById('collection-rule-body');
    tbody.innerHTML = '';

    if (rules.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align: center;">暂无采集规则数据</td></tr>';
        return;
    }

    rules.forEach(rule => {
        const row = document.createElement('tr');

        row.innerHTML = `
                    <td>${rule.id}</td>
                    <td>${rule.description}</td>
                    <td>${rule.ruleFuncCode}</td>
                    <td>${rule.startPoint}</td>
                    <td>${rule.endPoint}</td>
                    <td>
                        <button class="btn btn-primary btn-sm" onclick="editCollectionRule('${rule.id}')">
                            修改
                        </button>
                        <button class="btn btn-danger btn-sm" onclick="deleteCollectionRule('${rule.id}')">
                            删除
                        </button>
                    </td>
                `;

        tbody.appendChild(row);
    });
}

// 编辑采集规则
async function editCollectionRule(ruleId) {
    try {
        const response = await fetch(`/api/collection-rules/${ruleId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const rule = await response.json();

        // 填充表单
        document.getElementById('rule-id').value = rule.id;
        document.getElementById('rule-description').value = rule.description;
        document.getElementById('rule-func-code').value = rule.ruleFuncCode;
        document.getElementById('start-point').value = rule.startPoint;
        document.getElementById('end-point').value = rule.endPoint;

        // 更新标题
        document.getElementById('rule-edit-title').textContent = '编辑采集规则';

        // 显示模态框
        document.getElementById('rule-edit-modal').style.display = 'flex';
        document.getElementById('collection-rule-modal').style.display = 'none';
    } catch (error) {
        showNotification('获取采集规则详情失败' + error, 'error');
    }
}

// 保存采集规则
async function saveCollectionRule() {
    try {
        const formData = new FormData(document.getElementById('rule-edit-form'));
        const ruleData = {
            id: formData.get('rule-id') || undefined,
            description: formData.get('rule-description'),
            ruleFuncCode: parseInt(formData.get('rule-func-code')),
            startPoint: formData.get('start-point'),
            endPoint: formData.get('end-point'),
            deviceId: document.getElementById('device-select').value
        };

        const method = ruleData.id ? 'PUT' : 'POST';
        const url = ruleData.id ? `/api/collection-rules/${ruleData.id}` : '/api/collection-rules';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(ruleData)
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification('采集规则保存成功', 'success');
        document.getElementById('rule-edit-modal').style.display = 'none';
        document.getElementById('collection-rule-modal').style.display = 'flex';
        loadCollectionRules();
    } catch (error) {
        showNotification('保存采集规则失败' + error, 'error');
    }
}

// 删除采集规则
async function deleteCollectionRule(ruleId) {
    if (!confirm('确定要删除这个采集规则吗？')) return;

    try {
        const response = await fetch(`/api/collection-rules/${ruleId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification('采集规则删除成功', 'success');
        loadCollectionRules();
    } catch (error) {
        showNotification('删除采集规则失败' + error, 'error');
    }
}

// 切换设备状态
async function toggleDeviceStatus(deviceId, status) {
    try {
        const response = await fetch(`/api/devices/${deviceId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({status: status})
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification(`设备已${status ? '启动' : '停止'}`, 'success');
    } catch (error) {
        showNotification('切换设备状态失败' + error, 'error');
    }
}

// 保存设备
async function saveDevice() {
    try {
        const formData = new FormData(document.getElementById('device-form'));
        const deviceData = {
            id: formData.get('device-id'),
            name: formData.get('device-name'),
            code: formData.get('device-code'),
            table: formData.get('storage-table'),
            interfaceType: formData.get('interface-type'),
            address: formData.get('address'),
            baudRate: parseInt(formData.get('baudRate')),
            stopBits: parseInt(formData.get('stopBits')),
            dataBits: parseInt(formData.get('dataBits')),
            parity: formData.get('parity'),
            protocolType: formData.get('protocol-type'),
            deviceAddress: formData.get('device-address'),
            readTimeout: parseInt(formData.get('read-timeout')),
            writeTimeout: parseInt(formData.get('write-timeout')),
            status: formData.get('initial-status') === 'active'
        };

        const method = deviceData.id ? 'PUT' : 'POST';
        const url = deviceData.id ? `/api/devices/${deviceData.id}` : '/api/devices';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(deviceData)
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        showNotification('设备保存成功', 'success');
        document.getElementById('device-modal').style.display = 'none';
        loadDevices();
    } catch (error) {
        showNotification('保存设备失败' + error, 'error');
    }
}

// 保存点位
async function savePoint() {
    try {
        const formData = new FormData(document.getElementById('point-form'));
        const pointData = {
            id: formData.get('point-id') || undefined,
            functionCode: formData.get('function-code'),
            address: formData.get('point-address'),
            dataType: formData.get('data-type'),
            tag: formData.get('tag-flag'),
            luaExpression: formData.get('lua-expression'),
            description: formData.get('description'),
            alarmFlag: formData.get('alarm-flag'),
            alarmLevel: formData.get('alarm-level'),
            multiplier: parseFloat(formData.get('multiplier')) || 1,
            unit: formData.get('unit'),
            priority: parseInt(formData.get('priority')),
            endianness: formData.get('endianness'),
            bitCalculation: formData.get('bit-calculation'),
            startBit: parseInt(formData.get('start-bit')),
            endBit: parseInt(formData.get('end-bit')),
            storageMethod: formData.get('storage-method'),
            offset: parseInt(formData.get('offset')),
            store: parseInt(formData.get('store')),
            deviceId: document.getElementById('device-select').value
        };

        const method = pointData.id ? 'PUT' : 'POST';
        const url = pointData.id ? `/api/points/${pointData.id}` : '/api/points';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(pointData)
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const result = await response.json();
        showNotification('点位保存成功', 'success');
        document.getElementById('point-modal').style.display = 'none';
        loadPoints();
    } catch (error) {
        showNotification('保存点位失败' + error, 'error');
    }
}

// 编辑设备
async function editDevice(deviceId) {
    try {
        const response = await fetch(`/api/devices/${deviceId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const device = await response.json();

        // 填充表单
        document.getElementById('device-id').value = device.id;
        document.getElementById('device-name').value = device.name;
        document.getElementById('device-code').value = device.code;
        document.getElementById('storage-table').value = device.table;
        document.getElementById('interface-type').value = device.interfaceType;
        document.getElementById('address').value = device.address;
        document.getElementById('baudRate').value = device.baudRate;
        document.getElementById('stopBits').value = device.stopBits;
        document.getElementById('dataBits').value = device.dataBits;
        document.getElementById('parity').value = device.parity;
        document.getElementById('protocol-type').value = device.protocolType;
        document.getElementById('device-address').value = device.deviceAddress;
        document.getElementById('write-timeout').value = device.writeTimeout;
        document.getElementById('read-timeout').value = device.readTimeout;
        document.getElementById('initial-status').value = device.status ? 'active' : 'inactive';

        // 显示模态框
        document.getElementById('device-modal').style.display = 'flex';
    } catch (error) {
        showNotification('获取设备详情失败' + error, 'error');
    }
}

// 编辑点位
async function editPoint(pointId) {
    try {
        const response = await fetch(`/api/points/${pointId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const point = await response.json();

        // 填充表单
        document.getElementById('point-id').value = point.id;
        document.getElementById('function-code').value = point.functionCode;
        document.getElementById('point-address').value = point.address;
        document.getElementById('data-type').value = point.dataType;
        document.getElementById('tag-flag').value = point.tag;
        document.getElementById('lua-expression').value = point.luaExpression || '';
        document.getElementById('description').value = point.description;
        document.getElementById('alarm-flag').value = point.alarmFlag || '';
        document.getElementById('alarm-level').value = point.alarmLevel || 'None';
        document.getElementById('multiplier').value = point.multiplier || 1;
        document.getElementById('unit').value = point.unit || '';
        document.getElementById('priority').value = point.priority || '低';
        document.getElementById('endianness').value = point.endianness || '大端';
        document.getElementById('bit-calculation').value = point.bitCalculation || '整算';
        document.getElementById('start-bit').value = point.startBit || '';
        document.getElementById('end-bit').value = point.endBit || '';
        document.getElementById('storage-method').value = point.storageMethod || '';
        document.getElementById('offset').value = point.offset;
        document.getElementById('store').value = point.store || '';

        // 显示模态框
        document.getElementById('point-modal').style.display = 'flex';
    } catch (error) {
        showNotification('获取点位详情失败' + error, 'error');
    }
}

// 删除设备
async function deleteDevice(deviceId) {
    if (!confirm('确定要删除这个设备吗？')) return;

    try {
        const response = await fetch(`/api/devices/${deviceId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification('设备删除成功', 'success');
        loadDevices();
    } catch (error) {
        showNotification('删除设备失败' + error, 'error');
    }
}

// 删除点位
async function deletePoint(pointId) {
    if (!confirm('确定要删除这个点位吗？')) return;

    try {
        const response = await fetch(`/api/points/${pointId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification('点位删除成功', 'success');
        loadPoints();
    } catch (error) {
        showNotification('删除点位失败' + error, 'error');
    }
}

// 执行用户操作
async function executeUserAction(action) {
    try {
        let endpoint, method, body;

        switch (action) {
            case '暂停':
                endpoint = '/api/system/pause';
                method = 'POST';
                break;
            case '刷新':
                endpoint = '/api/system/flush';
                method = 'POST';
                break;
            case '清空':
                endpoint = '/api/data/clear';
                method = 'POST';
                break;
            case '导入':
                // 文件上传需要特殊处理
                const fileInput = document.createElement('input');
                fileInput.type = 'file';
                fileInput.accept = '.xlsx';
                fileInput.onchange = async (e) => {
                    const file = e.target.files[0];
                    if (!file) return;

                    const formData = new FormData();
                    formData.append('file', file);

                    try {
                        const response = await fetch('/api/config/import', {
                            method: 'POST',
                            body: formData
                        });

                        if (!response.ok) {
                            throw new Error(`HTTP error! status: ${response.status}`);
                        }

                        showNotification('配置文件导入成功', 'success');
                    } catch (error) {
                        showNotification('导入配置文件失败' + error, 'error');
                    }
                };
                fileInput.click();
                return;
            case '下载模板':
                // 直接下载模板文件
                window.open('/api/config/template', '_blank');
                showNotification('开始下载模板文件', 'info');
                return;
            default:
                showNotification('未知操作', 'error');
                return;
        }

        const response = await fetch(endpoint, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: body ? JSON.stringify(body) : undefined
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        showNotification(`${action}操作执行成功`, 'success');
    } catch (error) {
        showNotification(`执行${action}操作失败` + error, 'error');
    }
}

// 显示加载状态
function showLoading(tabId) {
    const tbody = document.querySelector(`#${tabId} tbody`);
    tbody.innerHTML = '<tr><td colspan="20" style="text-align: center;"><i class="fas fa-spinner fa-spin"></i> 加载中...</td></tr>';
}

// 显示通知
function showNotification(message, type = 'info') {
    // 创建或获取通知容器
    let container = document.getElementById('notification-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'notification-container';
        container.style.position = 'fixed';
        container.style.top = '20px';
        container.style.right = '20px';
        container.style.zIndex = '10000';
        document.body.appendChild(container);
    }

    // 创建通知元素
    const notification = document.createElement('div');
    notification.style.padding = '10px 15px';
    notification.style.marginBottom = '10px';
    notification.style.borderRadius = '4px';
    notification.style.color = 'white';
    notification.style.boxShadow = '0 2px 10px rgba(0,0,0,0.1)';
    notification.style.maxWidth = '300px';
    notification.style.wordBreak = 'break-word';

    // 设置背景颜色基于类型
    switch (type) {
        case 'success':
            notification.style.backgroundColor = '#2ecc71';
            break;
        case 'error':
            notification.style.backgroundColor = '#e74c3c';
            break;
        case 'warning':
            notification.style.backgroundColor = '#f39c12';
            break;
        default:
            notification.style.backgroundColor = '#3498db';
    }

    notification.textContent = message;
    container.appendChild(notification);

    // 3秒后自动移除
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 3000);
}

// 系统监控相关函数
// 更新统计信息
function updateStats(devices) {
    const totalDevices = devices.length;
    const onlineDevices = devices.filter(device => device.status === '在线').length;
    const alarmDevices = devices.filter(device => device.currentAlarmCount > 0).length;

    document.getElementById('totalDevices').textContent = totalDevices;
    document.getElementById('onlineDevices').textContent = onlineDevices;
    document.getElementById('alarmDevices').textContent = alarmDevices;
}

// 渲染设备表格（系统监控）
function renderDeviceTable(devices) {
    const deviceTableBody = document.getElementById('deviceTableBody');
    deviceTableBody.innerHTML = '';

    if (devices.length === 0) {
        deviceTableBody.innerHTML = `
                    <tr>
                        <td colspan="7" style="text-align: center; color: #718096;">
                            暂无设备数据
                        </td>
                    </tr>
                `;
        return;
    }

    devices.forEach(device => {
        const row = document.createElement('tr');

        // 格式化最后通讯时间
        const lastCommTime = new Date(device.lastCommunicationTime).toLocaleString();

        // 设置告警数量样式和点击事件
        let alarmCountClass = 'alarm-count no-alarm';
        let alarmCountClickable = '';
        if (device.currentAlarmCount > 0) {
            alarmCountClass = 'alarm-count has-alarm';
            alarmCountClickable = `onclick="showAlarmDetails(${device.id}, '${device.name}')"`;
        }

        row.innerHTML = `
                    <td>${device.id}</td>
                    <td>${device.name}</td>
                    <td>${device.code}</td>
                    <td>${device.totalPoints}</td>
                    <td>
                        <span class="${alarmCountClass}" ${alarmCountClickable}>
                            ${device.currentAlarmCount}
                        </span>
                    </td>
                    <td>
                        <span class="status ${device.status === '在线' ? 'online' : 'offline'}">
                            ${device.status}
                        </span>
                    </td>
                    <td>${lastCommTime}</td>
                `;

        deviceTableBody.appendChild(row);
    });
}

// 显示告警详情
window.showAlarmDetails = function (deviceId, deviceName) {
    const modalDeviceName = document.getElementById('modalDeviceName');
    const alarmModal = document.getElementById('alarmModal');
    const alarmTable = document.getElementById('alarmTable');
    const modalLoading = document.getElementById('modalLoading');
    const modalError = document.getElementById('modalError');

    modalDeviceName.textContent = deviceName;
    alarmModal.style.display = 'flex';
    alarmTable.style.display = 'none';
    modalError.style.display = 'none';
    modalLoading.style.display = 'flex';

    // 获取告警详情
    fetch(`/api/system/monitor?id=${deviceId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('获取告警详情失败');
            }
            return response.json();
        })
        .then(alarms => {
            modalLoading.style.display = 'none';

            if (alarms.length === 0) {
                document.getElementById('alarmTableBody').innerHTML = `
                            <tr>
                                <td colspan="5" style="text-align: center; color: #718096;">
                                    该设备暂无告警信息
                                </td>
                            </tr>
                        `;
                alarmTable.style.display = 'table';
                return;
            }

            renderAlarmTable(alarms);
            alarmTable.style.display = 'table';
        })
        .catch(error => {
            console.error('获取告警详情失败:', error);
            modalLoading.style.display = 'none';
            modalError.style.display = 'block';
        });
};

// 渲染告警表格
function renderAlarmTable(alarms) {
    const alarmTableBody = document.getElementById('alarmTableBody');
    alarmTableBody.innerHTML = '';

    alarms.forEach(alarm => {
        const row = document.createElement('tr');

        // 设置告警等级样式
        let levelClass = 'alarm-level ';
        if (alarm.level === '高') {
            levelClass += 'level-high';
        } else if (alarm.level === '中') {
            levelClass += 'level-medium';
        } else {
            levelClass += 'level-low';
        }

        row.innerHTML = `
                    <td>${alarm.point}</td>
                    <td>${alarm.description}</td>
                    <td>${alarm.currentValue}</td>
                    <td>
                        <span class="${levelClass}">
                            ${alarm.level}
                        </span>
                    </td>
                    <td>${alarm.condition}</td>
                `;

        alarmTableBody.appendChild(row);
    });
}

// 获取设备数据（系统监控）
function fetchDeviceData() {
    // 显示加载状态
    const refreshBtn = document.getElementById('refreshBtn');
    const deviceTableBody = document.getElementById('deviceTableBody');

    refreshBtn.classList.add('loading');
    deviceTableBody.innerHTML = `
                <tr>
                    <td colspan="7">
                        <div class="loading">
                            <div class="spinner"></div>
                        </div>
                    </td>
                </tr>
            `;

    // 发送POST请求获取设备数据
    fetch('/api/system/monitor', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('获取设备数据失败');
            }
            return response.json();
        })
        .then(devices => {
            updateStats(devices);
            renderDeviceTable(devices);
            refreshBtn.classList.remove('loading');
        })
        .catch(error => {
            deviceTableBody.innerHTML = `
                    <tr>
                        <td colspan="7" class="error-message">
                            加载设备数据失败，请<a href="javascript:void(0)" onclick="fetchDeviceData()">重试</a>
                        </td>
                    </tr>
                `;
            refreshBtn.classList.remove('loading');
        });
}