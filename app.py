"""
MachineServer - 统一仿真微服务平台
Main Application Entry Point
"""

from flask import Flask, jsonify
from flask_cors import CORS
from machineserver.api.simulation import simulation_bp
from machineserver.api.execution import execution_bp
from machineserver.api.coverage import coverage_bp
from machineserver.api.cosimulation import cosimulation_bp
from machineserver.utils.config import config
from machineserver.utils.logger import logger


def create_app():
    """Create and configure Flask application"""
    
    app = Flask(__name__)
    
    # Enable CORS
    CORS(app)
    
    # Register API blueprints
    app.register_blueprint(simulation_bp, url_prefix='/api/v1/simulation')
    app.register_blueprint(execution_bp, url_prefix='/api/v1/execution')
    app.register_blueprint(coverage_bp, url_prefix='/api/v1/coverage')
    app.register_blueprint(cosimulation_bp, url_prefix='/api/v1/cosimulation')
    
    # Root endpoint
    @app.route('/')
    def index():
        return jsonify({
            'name': 'MachineServer',
            'version': '0.1.0',
            'description': '统一仿真微服务平台 - A unified simulation microservice platform',
            'endpoints': {
                'simulation': '/api/v1/simulation',
                'execution': '/api/v1/execution',
                'coverage': '/api/v1/coverage',
                'cosimulation': '/api/v1/cosimulation'
            }
        })
    
    # Health check endpoint
    @app.route('/health')
    def health():
        return jsonify({'status': 'healthy'}), 200
    
    # API info endpoint
    @app.route('/api/v1')
    def api_info():
        return jsonify({
            'version': 'v1',
            'endpoints': {
                'simulation': {
                    'create': 'POST /api/v1/simulation/create',
                    'start': 'POST /api/v1/simulation/<id>/start',
                    'stop': 'POST /api/v1/simulation/<id>/stop',
                    'status': 'GET /api/v1/simulation/<id>/status',
                    'list': 'GET /api/v1/simulation/list',
                    'delete': 'DELETE /api/v1/simulation/<id>'
                },
                'execution': {
                    'load': 'POST /api/v1/execution/load',
                    'step': 'POST /api/v1/execution/<id>/step',
                    'run': 'POST /api/v1/execution/<id>/run',
                    'breakpoint': 'POST /api/v1/execution/<id>/breakpoint',
                    'registers': 'GET /api/v1/execution/<id>/registers',
                    'memory': 'GET /api/v1/execution/<id>/memory',
                    'status': 'GET /api/v1/execution/<id>/status'
                },
                'coverage': {
                    'start': 'POST /api/v1/coverage/<id>/start',
                    'stop': 'POST /api/v1/coverage/<id>/stop',
                    'report': 'GET /api/v1/coverage/<id>/report',
                    'export': 'GET /api/v1/coverage/<id>/export',
                    'status': 'GET /api/v1/coverage/<id>/status'
                },
                'cosimulation': {
                    'create': 'POST /api/v1/cosimulation/create',
                    'start': 'POST /api/v1/cosimulation/<id>/start',
                    'sync_step': 'POST /api/v1/cosimulation/<id>/sync-step',
                    'stop': 'POST /api/v1/cosimulation/<id>/stop',
                    'status': 'GET /api/v1/cosimulation/<id>/status',
                    'exchange': 'POST /api/v1/cosimulation/<id>/exchange'
                }
            }
        })
    
    # Error handlers
    @app.errorhandler(404)
    def not_found(error):
        return jsonify({'error': 'Not found'}), 404
    
    @app.errorhandler(500)
    def internal_error(error):
        return jsonify({'error': 'Internal server error'}), 500
    
    return app


def main():
    """Main entry point"""
    # Load configuration
    config.load_from_file('config.yaml')
    
    # Get server configuration
    host = config.get('server.host', '0.0.0.0')
    port = config.get('server.port', 5000)
    debug = config.get('server.debug', False)
    
    # Create and run application
    app = create_app()
    
    logger.info(f"Starting MachineServer on {host}:{port}")
    logger.info(f"Debug mode: {debug}")
    
    app.run(host=host, port=port, debug=debug)


if __name__ == '__main__':
    main()
