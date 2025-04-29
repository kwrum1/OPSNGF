import React, { useState } from 'react';
import { updateSuricataConfig } from '@/api/suricata';

const SuricataSetting: React.FC = () => {
  const [config, setConfig] = useState({});

  const handleSubmit = async () => {
    await updateSuricataConfig(config);
    alert('配置已更新');
  };

  return (
    <div>
      <h1>Suricata 配置</h1>
      {/* 在此添加配置表单 */}
      <button onClick={handleSubmit}>保存配置</button>
    </div>
  );
};

export default SuricataSetting;
