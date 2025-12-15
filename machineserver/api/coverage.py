"""Coverage API routes"""

from flask import Blueprint, request, jsonify
from ..core.coverage_analyzer import CoverageAnalyzer

coverage_bp = Blueprint('coverage', __name__)
coverage_analyzer = CoverageAnalyzer()


@coverage_bp.route('/<execution_id>/start', methods=['POST'])
def start_coverage(execution_id):
    """Start code coverage collection"""
    result = coverage_analyzer.start_coverage(execution_id)
    return jsonify(result), 201


@coverage_bp.route('/<session_id>/stop', methods=['POST'])
def stop_coverage(session_id):
    """Stop code coverage collection"""
    result = coverage_analyzer.stop_coverage(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@coverage_bp.route('/<session_id>/report', methods=['GET'])
def get_coverage_report(session_id):
    """Get coverage report"""
    result = coverage_analyzer.get_report(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@coverage_bp.route('/<session_id>/export', methods=['GET'])
def export_coverage(session_id):
    """Export coverage data"""
    format = request.args.get('format', 'json')
    
    result = coverage_analyzer.export_coverage(session_id, format)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200


@coverage_bp.route('/<session_id>/status', methods=['GET'])
def get_coverage_status(session_id):
    """Get coverage session status"""
    result = coverage_analyzer.get_status(session_id)
    
    if 'error' in result:
        return jsonify(result), 404
    
    return jsonify(result), 200
