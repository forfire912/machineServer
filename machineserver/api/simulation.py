"""Simulation API routes"""

from flask import Blueprint, request, jsonify
from ..core.simulation_manager import SimulationManager

simulation_bp = Blueprint('simulation', __name__)
sim_manager = SimulationManager()


@simulation_bp.route('/create', methods=['POST'])
def create_simulation():
    """Create a new simulation instance"""
    data = request.get_json()
    
    if not data or 'processor_type' not in data:
        return jsonify({'error': 'processor_type is required'}), 400
    
    processor_type = data['processor_type']
    config = data.get('config', {})
    
    result = sim_manager.create_simulation(processor_type, config)
    return jsonify(result), 201


@simulation_bp.route('/<simulation_id>/start', methods=['POST'])
def start_simulation(simulation_id):
    """Start a simulation instance"""
    result = sim_manager.start_simulation(simulation_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@simulation_bp.route('/<simulation_id>/stop', methods=['POST'])
def stop_simulation(simulation_id):
    """Stop a simulation instance"""
    result = sim_manager.stop_simulation(simulation_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@simulation_bp.route('/<simulation_id>/status', methods=['GET'])
def get_simulation_status(simulation_id):
    """Get simulation status"""
    result = sim_manager.get_status(simulation_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@simulation_bp.route('/list', methods=['GET'])
def list_simulations():
    """List all simulation instances"""
    result = sim_manager.list_simulations()
    return jsonify(result), 200


@simulation_bp.route('/<simulation_id>', methods=['DELETE'])
def delete_simulation(simulation_id):
    """Delete a simulation instance"""
    result = sim_manager.delete_simulation(simulation_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200
