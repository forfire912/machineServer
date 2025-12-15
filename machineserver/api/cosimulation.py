"""Co-Simulation API routes"""

from flask import Blueprint, request, jsonify
from ..core.cosim_coordinator import CoSimCoordinator

cosimulation_bp = Blueprint('cosimulation', __name__)
cosim_coordinator = CoSimCoordinator()


@cosimulation_bp.route('/create', methods=['POST'])
def create_cosimulation():
    """Create a co-simulation session"""
    data = request.get_json()
    
    if not data or 'components' not in data:
        return jsonify({'error': 'components list is required'}), 400
    
    components = data['components']
    
    result = cosim_coordinator.create_cosimulation(components)
    return jsonify(result), 201


@cosimulation_bp.route('/<session_id>/start', methods=['POST'])
def start_cosimulation(session_id):
    """Start co-simulation"""
    result = cosim_coordinator.start_cosimulation(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@cosimulation_bp.route('/<session_id>/sync-step', methods=['POST'])
def sync_step(session_id):
    """Execute synchronized step"""
    data = request.get_json() or {}
    time_step_ns = data.get('time_step_ns', 1000)
    
    result = cosim_coordinator.sync_step(session_id, time_step_ns)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@cosimulation_bp.route('/<session_id>/stop', methods=['POST'])
def stop_cosimulation(session_id):
    """Stop co-simulation"""
    result = cosim_coordinator.stop_cosimulation(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@cosimulation_bp.route('/<session_id>/status', methods=['GET'])
def get_cosimulation_status(session_id):
    """Get co-simulation status"""
    result = cosim_coordinator.get_status(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@cosimulation_bp.route('/<session_id>/exchange', methods=['POST'])
def exchange_data(session_id):
    """Exchange data between components"""
    data = request.get_json()
    
    if not data or 'source' not in data or 'target' not in data:
        return jsonify({'error': 'source and target are required'}), 400
    
    source = data['source']
    target = data['target']
    exchange_data = data.get('data', {})
    
    result = cosim_coordinator.exchange_data(session_id, source, target, exchange_data)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200
