package main

// HTML template for the visualization page
var graphVisualizationHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Identity Graph Visualization</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            display: flex;
            flex-direction: column;
            height: 100vh;
        }
        #controls {
            padding: 15px;
            background-color: #f8f9fa;
            border-bottom: 1px solid #dee2e6;
        }
        #graph-container {
            flex-grow: 1;
            overflow: hidden;
            position: relative;
        }
        #graph {
            width: 100%;
            height: 100%;
        }
        .node {
            cursor: pointer;
        }
        .link {
            stroke-opacity: 0.6;
        }
        .node text {
            font-size: 12px;
            pointer-events: none;
        }
        .node-detail {
            position: absolute;
            right: 20px;
            top: 20px;
            background: white;
            border: 1px solid #ccc;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.2);
            max-width: 300px;
            display: none;
        }
        .controls-form {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            align-items: flex-end;
        }
        .form-group {
            display: flex;
            flex-direction: column;
        }
        label {
            font-size: 14px;
            margin-bottom: 5px;
        }
        button {
            padding: 7px 15px;
            background: #4CAF50;
            color: white;
            border: none;
            cursor: pointer;
            border-radius: 3px;
        }
        button:hover {
            background: #45a049;
        }
        select, input {
            padding: 6px;
            border: 1px solid #ddd;
            border-radius: 3px;
        }
        .legend {
            position: absolute;
            left: 20px;
            bottom: 20px;
            background: rgba(255, 255, 255, 0.9);
            padding: 10px;
            border-radius: 5px;
            border: 1px solid #ccc;
        }
        .legend-item {
            display: flex;
            align-items: center;
            margin-bottom: 5px;
        }
        .legend-color {
            width: 15px;
            height: 15px;
            margin-right: 8px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <div id="controls">
        <h2>Identity Graph Visualization</h2>
        <div class="controls-form">
            <div class="form-group">
                <label for="entity-type">Entity Type:</label>
                <select id="entity-type">
                    <option value="">All</option>
                    <option value="user">User</option>
                    <option value="organization">Organization</option>
                    <option value="project">Project</option>
                    <option value="group">Group</option>
                    <option value="task">Task</option>
                </select>
            </div>
            <div class="form-group">
                <label for="entity-id">Entity ID:</label>
                <input type="text" id="entity-id" placeholder="e.g., alice">
            </div>
            <div class="form-group">
                <label for="relation-type">Relation Type:</label>
                <input type="text" id="relation-type" placeholder="e.g., owner">
            </div>
            <div class="form-group">
                <label for="depth">Depth:</label>
                <select id="depth">
                    <option value="1">1</option>
                    <option value="2" selected>2</option>
                    <option value="3">3</option>
                    <option value="4">4</option>
                    <option value="5">5</option>
                </select>
            </div>
            <div class="form-group">
                <button id="refresh-graph">Refresh Graph</button>
            </div>
        </div>
    </div>
    <div id="graph-container">
        <svg id="graph"></svg>
        <div id="node-detail" class="node-detail"></div>
        <div id="graph-legend" class="legend"></div>
    </div>

    <script src="https://d3js.org/d3.v7.min.js"></script>
    <script>
        // Graph visualization settings
        const width = window.innerWidth;
        const height = window.innerHeight - document.getElementById('controls').offsetHeight;
        
        // Color scale for entity types
        const entityColors = {
            'user': '#4285F4',
            'organization': '#EA4335',
            'group': '#FBBC05',
            'project': '#34A853',
            'task': '#FF6D01',
            'default': '#9AA0A6'
        };
        
        // Create SVG
        const svg = d3.select('#graph')
            .attr('width', width)
            .attr('height', height);
            
        // Create zoom behavior
        const zoom = d3.zoom()
            .scaleExtent([0.1, 8])
            .on('zoom', (event) => {
                g.attr('transform', event.transform);
            });
            
        svg.call(zoom);
        
        // Create main group for zooming
        const g = svg.append('g');
        
        // Create arrow markers for directed links
        svg.append('defs').selectAll('marker')
            .data(['default', 'owner', 'admin', 'member', 'manager'])
            .enter().append('marker')
            .attr('id', d => ` + "`arrow-${d}`" + `)
            .attr('viewBox', '0 -5 10 10')
            .attr('refX', 20)
            .attr('refY', 0)
            .attr('markerWidth', 6)
            .attr('markerHeight', 6)
            .attr('orient', 'auto')
            .append('path')
            .attr('d', 'M0,-5L10,0L0,5')
            .attr('fill', d => d === 'default' ? '#999' : d3.color(getRelationColor(d)).darker());
        
        // Create link and node groups
        const linksGroup = g.append('g').attr('class', 'links');
        const nodesGroup = g.append('g').attr('class', 'nodes');
        
        // Force simulation
        let simulation = d3.forceSimulation()
            .force('link', d3.forceLink().id(d => d.id).distance(150))
            .force('charge', d3.forceManyBody().strength(-300))
            .force('center', d3.forceCenter(width / 2, height / 2))
            .force('collision', d3.forceCollide().radius(50));
        
        // Function to get color based on entity type
        function getEntityColor(type) {
            return entityColors[type] || entityColors.default;
        }
        
        // Function to get color based on relation type
        function getRelationColor(type) {
            const relationColors = {
                'owner': '#AA00FF',
                'admin': '#00C853',
                'member': '#2979FF',
                'manager': '#FF6D00',
                'contributor': '#FFC400',
                'default': '#9E9E9E'
            };
            return relationColors[type] || relationColors.default;
        }
        
        // Function to refresh graph data
        function refreshGraph() {
            const entityType = document.getElementById('entity-type').value;
            const entityId = document.getElementById('entity-id').value;
            const relationType = document.getElementById('relation-type').value;
            const depth = document.getElementById('depth').value;
            
            // Build API URL with filters
            const url = '/api/graph?' + new URLSearchParams({
                entity_type: entityType,
                entity_id: entityId,
                relation_type: relationType,
                depth: depth
            });
            
            // Fetch graph data
            fetch(url)
                .then(response => response.json())
                .then(data => updateVisualization(data))
                .catch(error => console.error('Error fetching graph data:', error));
        }
        
        // Function to update visualization with new data
        function updateVisualization(data) {
            // Stop existing simulation
            if (simulation) simulation.stop();
            
            // Clear existing elements
            linksGroup.selectAll('*').remove();
            nodesGroup.selectAll('*').remove();
            
            // If no data, show message
            if (data.nodes.length === 0) {
                svg.append('text')
                    .attr('x', width / 2)
                    .attr('y', height / 2)
                    .attr('text-anchor', 'middle')
                    .text('No data found for the selected filters.');
                return;
            }
            
            // Create links
            const link = linksGroup.selectAll('.link')
                .data(data.links)
                .enter().append('g')
                .attr('class', 'link');
                
            // Link lines
            link.append('line')
                .attr('stroke', d => getRelationColor(d.type))
                .attr('stroke-width', 2)
                .attr('marker-end', d => ` + "`url(#arrow-${d.type || 'default'})`" + `);
                
            // Link labels
            link.append('text')
                .attr('dy', -5)
                .attr('text-anchor', 'middle')
                .attr('fill', d => d3.color(getRelationColor(d.type)).darker())
                .text(d => d.label);
            
            // Create nodes
            const node = nodesGroup.selectAll('.node')
                .data(data.nodes)
                .enter().append('g')
                .attr('class', 'node')
                .call(d3.drag()
                    .on('start', dragstarted)
                    .on('drag', dragged)
                    .on('end', dragended))
                .on('click', showNodeDetails);
                
            // Node circles
            node.append('circle')
                .attr('r', 10)
                .attr('fill', d => getEntityColor(d.group));
                
            // Node labels
            node.append('text')
                .attr('dx', 15)
                .attr('dy', 5)
                .text(d => d.label)
                .attr('fill', '#333');
                
            // Create legend
            updateLegend(data);
                
            // Update simulation
            simulation = d3.forceSimulation(data.nodes)
                .force('link', d3.forceLink(data.links).id(d => d.id).distance(150))
                .force('charge', d3.forceManyBody().strength(-300))
                .force('center', d3.forceCenter(width / 2, height / 2))
                .force('collision', d3.forceCollide().radius(50))
                .on('tick', ticked);
                
            // Handle tick events
            function ticked() {
                link.select('line')
                    .attr('x1', d => d.source.x)
                    .attr('y1', d => d.source.y)
                    .attr('x2', d => d.target.x)
                    .attr('y2', d => d.target.y);
                    
                link.select('text')
                    .attr('x', d => (d.source.x + d.target.x) / 2)
                    .attr('y', d => (d.source.y + d.target.y) / 2);
                    
                node.attr('transform', d => ` + "`translate(${d.x},${d.y})`" + `);
            }
            
            // Drag functions
            function dragstarted(event, d) {
                if (!event.active) simulation.alphaTarget(0.3).restart();
                d.fx = d.x;
                d.fy = d.y;
            }
            
            function dragged(event, d) {
                d.fx = event.x;
                d.fy = event.y;
            }
            
            function dragended(event, d) {
                if (!event.active) simulation.alphaTarget(0);
                d.fx = null;
                d.fy = null;
            }
        }
        
        // Function to update legend
        function updateLegend(data) {
            const legend = document.getElementById('graph-legend');
            legend.innerHTML = '<h3>Legend</h3>';
            
            // Entity types legend
            const entityTypes = [...new Set(data.nodes.map(node => node.type))];
            legend.innerHTML += '<div><strong>Entity Types:</strong></div>';
            
            entityTypes.forEach(type => {
                const item = document.createElement('div');
                item.className = 'legend-item';
                
                const color = document.createElement('div');
                color.className = 'legend-color';
                color.style.backgroundColor = getEntityColor(type);
                
                item.appendChild(color);
                item.appendChild(document.createTextNode(type));
                legend.appendChild(item);
            });
            
            // Relation types legend
            const relationTypes = [...new Set(data.links.map(link => link.type))];
            legend.innerHTML += '<div style="margin-top: 10px"><strong>Relation Types:</strong></div>';
            
            relationTypes.forEach(type => {
                const item = document.createElement('div');
                item.className = 'legend-item';
                
                const color = document.createElement('div');
                color.className = 'legend-color';
                color.style.backgroundColor = getRelationColor(type);
                
                item.appendChild(color);
                item.appendChild(document.createTextNode(type));
                legend.appendChild(item);
            });
        }
        
        // Function to show node details
        function showNodeDetails(event, d) {
            const detailPanel = document.getElementById('node-detail');
            detailPanel.style.display = 'block';
            
            let html = ` + "`" + `
                <h3>${d.label}</h3>
                <p><strong>ID:</strong> ${d.id}</p>
                <p><strong>Type:</strong> ${d.type}</p>
            ` + "`" + `;
            
            // Get related entities
            const relatedLinks = [...linksGroup.selectAll('line')]
                .map(line => d3.select(line).datum())
                .filter(link => link.source.id === d.id || link.target.id === d.id);
                
            if (relatedLinks.length > 0) {
                html += '<p><strong>Related Entities:</strong></p><ul>';
                
                relatedLinks.forEach(link => {
                    const isSource = link.source.id === d.id;
                    const relatedNode = isSource ? link.target : link.source;
                    const direction = isSource ? 'outgoing' : 'incoming';
                    
                    html += ` + "`<li>${direction} <strong>${link.type}</strong> relation with ${relatedNode.label} (${relatedNode.id})</li>`" + `; 
                });
                
                html += '</ul>';
            }
            
            detailPanel.innerHTML = html;
        }
        
        // Add event listener for refresh button
        document.getElementById('refresh-graph').addEventListener('click', refreshGraph);
        
        // Close node details when clicking elsewhere
        document.addEventListener('click', function(event) {
            const detailPanel = document.getElementById('node-detail');
            const isClickInsideNode = event.target.closest('.node');
            const isClickInsidePanel = event.target.closest('#node-detail');
            
            if (!isClickInsideNode && !isClickInsidePanel) {
                detailPanel.style.display = 'none';
            }
        });
        
        // Initialize graph on page load
        window.addEventListener('load', refreshGraph);
        
        // Handle window resize
        window.addEventListener('resize', function() {
            const newWidth = window.innerWidth;
            const newHeight = window.innerHeight - document.getElementById('controls').offsetHeight;
            
            svg.attr('width', newWidth)
               .attr('height', newHeight);
               
            simulation.force('center', d3.forceCenter(newWidth / 2, newHeight / 2))
                       .restart();
        });

        const permissionPathUI = ` + "`<div id=\"permission-path-ui\" style=\"display: none; margin-top: 20px;\">" + `
    <h3>Permission Path Analysis</h3>
    <div class="controls-form">
        <div class="form-group">
            <label for="subject-type">Subject Type:</label>
            <select id="subject-type">
                <option value="user">User</option>
                <option value="organization">Organization</option>
                <option value="group">Group</option>
                <option value="project">Project</option>
            </select>
        </div>
        <div class="form-group">
            <label for="subject-id">Subject ID:</label>
            <input type="text" id="subject-id" placeholder="e.g., alice">
        </div>
        <div class="form-group">
            <label for="permission-name">Permission:</label>
            <input type="text" id="permission-name" placeholder="e.g., manage_settings">
        </div>
        <div class="form-group">
            <label for="object-type">Object Type:</label>
            <select id="object-type">
                <option value="organization">Organization</option>
                <option value="user">User</option>
                <option value="group">Group</option>
                <option value="project">Project</option>
            </select>
        </div>
        <div class="form-group">
            <label for="object-id">Object ID:</label>
            <input type="text" id="object-id" placeholder="e.g., acme">
        </div>
        <div class="form-group">
            <button id="analyze-permission">Analyze Permission Path</button>
        </div>
    </div>
    <div id="permission-result" style="margin-top: 15px; display: none;">
        <div id="permission-status" style="padding: 10px; border-radius: 5px; margin-bottom: 10px;"></div>
        <div id="permission-expression" style="font-family: monospace; background: #f5f5f5; padding: 10px; border-radius: 5px; margin-bottom: 10px;"></div>
        <div id="permission-paths-count"></div>
        <div id="permission-paths-list" style="margin-top: 10px;"></div>
    </div>
</div>` + "`" + `;

// Add this to the controls div at the end of the existing content
document.getElementById('controls').innerHTML += permissionPathUI;

// Add UI toggle button
const controlsForm = document.querySelector('.controls-form');
const toggleButton = document.createElement('div');
toggleButton.className = 'form-group';
toggleButton.innerHTML = '<button id="toggle-permission-ui">Permission Path Analysis</button>';
controlsForm.appendChild(toggleButton);

// Toggle UI visibility
document.getElementById('toggle-permission-ui').addEventListener('click', function() {
    const permissionUI = document.getElementById('permission-path-ui');
    if (permissionUI.style.display === 'none') {
        permissionUI.style.display = 'block';
        this.textContent = 'Hide Permission Analysis';
    } else {
        permissionUI.style.display = 'none';
        this.textContent = 'Permission Path Analysis';
    }
});

// Handle permission path analysis
document.getElementById('analyze-permission').addEventListener('click', function() {
    const subjectType = document.getElementById('subject-type').value;
    const subjectId = document.getElementById('subject-id').value;
    const permission = document.getElementById('permission-name').value;
    const objectType = document.getElementById('object-type').value;
    const objectId = document.getElementById('object-id').value;
    
    // Validate inputs
    if (!subjectType || !subjectId || !permission || !objectType || !objectId) {
        alert('All fields are required');
        return;
    }
    
    // Make API request
    fetch('/api/permission-path', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            subject_type: subjectType,
            subject_id: subjectId,
            permission: permission,
            object_type: objectType,
            object_id: objectId
        }),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Permission path analysis failed');
        }
        return response.json();
    })
    .then(data => {
        // Show results section
        document.getElementById('permission-result').style.display = 'block';
        
        // Update permission status
        const statusElement = document.getElementById('permission-status');
        if (data.allowed) {
            statusElement.textContent = '✅ Permission Granted';
            statusElement.style.backgroundColor = '#d4edda';
            statusElement.style.color = '#155724';
        } else {
            statusElement.textContent = '❌ Permission Denied';
            statusElement.style.backgroundColor = '#f8d7da';
            statusElement.style.color = '#721c24';
        }
        
        // Show permission expression
        document.getElementById('permission-expression').textContent = 'Expression: ' + data.expression;
        
        // Show paths count
        const pathsCountElement = document.getElementById('permission-paths-count');
        if (data.paths && data.paths.length > 0) {
            pathsCountElement.textContent = ` + "`Found ${data.paths.length} possible permission path(s):`" + `;
            
            // Generate paths list
            const pathsList = document.getElementById('permission-paths-list');
            pathsList.innerHTML = '';
            
            data.paths.forEach((path, index) => {
                const pathItem = document.createElement('div');
                pathItem.style.marginBottom = '10px';
                pathItem.style.padding = '10px';
                pathItem.style.backgroundColor = '#f8f9fa';
                pathItem.style.borderRadius = '5px';
                pathItem.style.cursor = 'pointer';
                
                // Create path description
                let pathDescription = ` + "`<strong>Path ${index + 1}:</strong> `" + `;
                
                path.forEach((link, linkIndex) => {
                    const sourceId = link.source.includes(':') ? link.source.split(':')[1] : link.source;
                    const targetId = link.target.includes(':') ? link.target.split(':')[1] : link.target;
                    
                    pathDescription += sourceId;
                    pathDescription += ` + "` <span style=\"color: #007bff;\">${link.type}</span> `" + `;
                    
                    if (linkIndex === path.length - 1) {
                        pathDescription += targetId;
                    }
                });
                
                pathItem.innerHTML = pathDescription;
                
                // Add click handler to highlight this path in the graph
                pathItem.addEventListener('click', () => highlightPath(data.nodes, path));
                
                pathsList.appendChild(pathItem);
            });
        } else {
            pathsCountElement.textContent = 'No permission paths found.';
            document.getElementById('permission-paths-list').innerHTML = '';
        }
        
        // Visualize the graph with these nodes and links
        updateVisualization({
            nodes: data.nodes,
            links: data.links
        });
    })
    .catch(error => {
        console.error('Error:', error);
        alert('Failed to analyze permission path: ' + error.message);
    });
});

// Function to highlight a specific path in the graph
function highlightPath(nodes, path) {
    // Reset all nodes and links
    d3.selectAll('.node circle').attr('stroke-width', 0);
    d3.selectAll('.link line').attr('stroke-opacity', 0.6).attr('stroke-width', 2);
    
    // Highlight nodes in the path
    const nodeIds = new Set();
    path.forEach(link => {
        nodeIds.add(link.source);
        nodeIds.add(link.target);
    });
    
    d3.selectAll('.node')
        .filter(d => nodeIds.has(d.id))
        .select('circle')
        .attr('stroke', '#ff0000')
        .attr('stroke-width', 3);
    
    // Highlight links in the path
    const linkMap = new Map(path.map(link => [link.source + '-' + link.target, true]));
    
    d3.selectAll('.link line')
        .attr('stroke-opacity', function(d) {
            const key = d.source.id + '-' + d.target.id;
            return linkMap.has(key) ? 1 : 0.2;
        })
        .attr('stroke-width', function(d) {
            const key = d.source.id + '-' + d.target.id;
            return linkMap.has(key) ? 4 : 2;
        });
}
    </script>
</body>
</html>
`
