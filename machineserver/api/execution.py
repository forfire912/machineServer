"""Execution API routes"""

from flask import Blueprint, request, jsonify
from ..core.execution_engine import ExecutionEngine

execution_bp = Blueprint('execution', __name__)
exec_engine = ExecutionEngine()


@execution_bp.route('/load', methods=['POST'])
def load_program():
    """Load a program for execution"""
    data = request.get_json()
    
    if not data or 'simulation_id' not in data or 'program_path' not in data:
        return jsonify({'error': 'simulation_id and program_path are required'}), 400
    
    simulation_id = data['simulation_id']
    program_path = data['program_path']
    
    result = exec_engine.load_program(simulation_id, program_path)
    return jsonify(result), 201


@execution_bp.route('/<session_id>/step', methods=['POST'])
def step_execution(session_id):
    """Execute instruction steps"""
    data = request.get_json() or {}
    count = data.get('count', 1)
    
    result = exec_engine.step(session_id, count)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/run', methods=['POST'])
def run_execution(session_id):
    """Run program until breakpoint or completion"""
    result = exec_engine.run(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/breakpoint', methods=['POST'])
def set_breakpoint(session_id):
    """Set a breakpoint"""
    data = request.get_json()
    
    if not data or 'address' not in data:
        return jsonify({'error': 'address is required'}), 400
    
    address = data['address']
    result = exec_engine.set_breakpoint(session_id, address)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/breakpoint/<address>', methods=['DELETE'])
def remove_breakpoint(session_id, address):
    """Remove a breakpoint"""
    result = exec_engine.remove_breakpoint(session_id, address)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/registers', methods=['GET'])
def read_registers(session_id):
    """Read processor registers"""
    result = exec_engine.read_registers(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/memory', methods=['GET'])
def read_memory(session_id):
    """Read memory contents"""
    address = request.args.get('address', '0x00000000')
    size = int(request.args.get('size', 256))
    
    result = exec_engine.read_memory(session_id, address, size)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@execution_bp.route('/<session_id>/status', methods=['GET'])
def get_execution_status(session_id):
    """Get execution session status"""
    result = exec_engine.get_status(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200
