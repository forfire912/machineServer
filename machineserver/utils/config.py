"""Configuration management module"""

import os
import yaml
from typing import Dict, Any


class Config:
    """Configuration manager for MachineServer"""
    
    DEFAULT_CONFIG = {
        'server': {
            'host': '0.0.0.0',
            'port': 5000,
            'debug': False
        },
        'simulation': {
            'max_instances': 10,
            'timeout': 3600
        },
        'logging': {
            'level': 'INFO',
            'file': 'machineserver.log'
        }
    }
    
    def __init__(self, config_path: str = None):
        """Initialize configuration
        
        Args:
            config_path: Path to YAML configuration file
        """
        self.config = self.DEFAULT_CONFIG.copy()
        
        if config_path and os.path.exists(config_path):
            self.load_from_file(config_path)
    
    def load_from_file(self, config_path: str) -> None:
        """Load configuration from YAML file
        
        Args:
            config_path: Path to YAML configuration file
        """
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                file_config = yaml.safe_load(f)
                if file_config:
                    self._merge_config(file_config)
        except Exception as e:
            print(f"Warning: Failed to load config from {config_path}: {e}")
    
    def _merge_config(self, new_config: Dict[str, Any]) -> None:
        """Merge new configuration with existing
        
        Args:
            new_config: New configuration dictionary
        """
        for key, value in new_config.items():
            if key in self.config and isinstance(value, dict):
                self.config[key].update(value)
            else:
                self.config[key] = value
    
    def get(self, key: str, default: Any = None) -> Any:
        """Get configuration value
        
        Args:
            key: Configuration key (supports dot notation, e.g., 'server.port')
            default: Default value if key not found
            
        Returns:
            Configuration value
        """
        keys = key.split('.')
        value = self.config
        
        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                return default
        
        return value
    
    def set(self, key: str, value: Any) -> None:
        """Set configuration value
        
        Args:
            key: Configuration key (supports dot notation)
            value: Value to set
        """
        keys = key.split('.')
        config = self.config
        
        for k in keys[:-1]:
            if k not in config:
                config[k] = {}
            config = config[k]
        
        config[keys[-1]] = value


# Global configuration instance
config = Config()
