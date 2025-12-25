import React, { useState, useEffect } from 'react';
import '../styles/AdminKnowledgePage.css';
import axios from 'axios';

interface KnowledgeFormData {
  title: string;
  content: string;
  category: string;
  difficulty: string;
  tags: string[];
  language: string;
}

interface KnowledgeItem {
  id: string;
  title: string;
  content: string;
  category: string;
  difficulty: string;
  tags: string[];
  language: string;
  created_at: string;
}

const AdminKnowledgePage: React.FC = () => {
  const [formData, setFormData] = useState<KnowledgeFormData>({
    title: '',
    content: '',
    category: 'stain_removal',
    difficulty: 'basic',
    tags: [],
    language: 'KR',
  });

  const [tagInput, setTagInput] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [knowledgeList, setKnowledgeList] = useState<KnowledgeItem[]>([]);
  const [isLoadingList, setIsLoadingList] = useState(true);

  // Fetch knowledge list on mount
  useEffect(() => {
    fetchKnowledgeList();
  }, []);

  const fetchKnowledgeList = async () => {
    try {
      setIsLoadingList(true);
      const response = await axios.get('/api/knowledge');
      setKnowledgeList(response.data.items || []);
    } catch (error) {
      console.error('Failed to fetch knowledge list:', error);
    } finally {
      setIsLoadingList(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('정말 삭제하시겠습니까?')) return;
    
    try {
      await axios.delete(`/api/knowledge/${id}`);
      setMessage({ type: 'success', text: '삭제되었습니다.' });
      fetchKnowledgeList();
    } catch (error) {
      setMessage({ type: 'error', text: '삭제에 실패했습니다.' });
    }
  };

  const categories = [
    { value: 'stain_removal', label: '얼룩제거' },
    { value: 'fabric_understanding', label: '원단이해' },
    { value: 'accident_prevention', label: '세탁사고와 방지' },
    { value: 'laundry_technique', label: '세탁기술' },
    { value: 'equipment_operation', label: '세탁장비운영' },
    { value: 'marketing', label: '마케팅과 고객관리' },
    { value: 'others', label: '기타' },
  ];

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleAddTag = () => {
    if (tagInput.trim() && !formData.tags.includes(tagInput.trim())) {
      setFormData(prev => ({
        ...prev,
        tags: [...prev.tags, tagInput.trim()],
      }));
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setFormData(prev => ({
      ...prev,
      tags: prev.tags.filter(tag => tag !== tagToRemove),
    }));
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    setSelectedFiles(prev => [...prev, ...files]);
  };

  const handleRemoveFile = (index: number) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
  };

  const extractTextFromFiles = async (): Promise<string> => {
    let extractedText = '';

    for (const file of selectedFiles) {
      const formData = new FormData();
      formData.append('file', file);

      try {
        const response = await axios.post('/api/extract-text', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        extractedText += response.data.text + '\n\n';
      } catch (error) {
        console.error(`Failed to extract text from ${file.name}:`, error);
      }
    }

    return extractedText;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setMessage(null);

    try {
      let finalContent = formData.content;

      // Extract text from uploaded files
      if (selectedFiles.length > 0) {
        const extractedText = await extractTextFromFiles();
        finalContent = formData.content + '\n\n' + extractedText;
      }

      const payload = {
        ...formData,
        content: finalContent,
      };

      const response = await axios.post('/api/knowledge', payload, {
        headers: { 'Content-Type': 'application/json' },
      });

      setMessage({ type: 'success', text: '지식이 성공적으로 등록되었습니다!' });

      // Reset form
      setFormData({
        title: '',
        content: '',
        category: 'stain_removal',
        difficulty: 'basic',
        tags: [],
        language: 'KR',
      });
      setSelectedFiles([]);
      setTagInput('');
      
      // Refresh the knowledge list
      fetchKnowledgeList();
    } catch (error: any) {
      setMessage({
        type: 'error',
        text: error.response?.data?.message || '지식 등록에 실패했습니다.'
      });
    } finally {
      setIsLoading(false);
    }
  };

  const getFileIcon = (fileName: string) => {
    const ext = fileName.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'pdf':
        return '📄';
      case 'txt':
      case 'md':
        return '📝';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'webp':
        return '🖼️';
      default:
        return '📎';
    }
  };

  return (
    <div className="admin-knowledge-page">
      <div className="admin-container">
        <div className="admin-header">
          <div className="header-nav">
            <a href="/" className="nav-link">고객 페이지로 이동</a>
          </div>
          <h1>지식 관리</h1>
          <p>세탁 지식 데이터베이스에 새로운 지식을 추가합니다</p>
        </div>

        {message && (
          <div className={`message ${message.type}`}>
            {message.text}
          </div>
        )}

        <form className="knowledge-form" onSubmit={handleSubmit}>
          {/* Title */}
          <div className="form-group">
            <label htmlFor="title">제목 *</label>
            <input
              type="text"
              id="title"
              name="title"
              value={formData.title}
              onChange={handleInputChange}
              placeholder="예: 커피 얼룩 제거 방법"
              required
            />
          </div>

          {/* Category */}
          <div className="form-row">
            <div className="form-group">
              <label htmlFor="category">카테고리 *</label>
              <select
                id="category"
                name="category"
                value={formData.category}
                onChange={handleInputChange}
                required
              >
                {categories.map(cat => (
                  <option key={cat.value} value={cat.value}>
                    {cat.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="difficulty">난이도 *</label>
              <select
                id="difficulty"
                name="difficulty"
                value={formData.difficulty}
                onChange={handleInputChange}
                required
              >
                <option value="basic">Basic (기본)</option>
                <option value="expert">Expert (전문가)</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="language">언어</label>
              <select
                id="language"
                name="language"
                value={formData.language}
                onChange={handleInputChange}
              >
                <option value="KR">한국어</option>
                <option value="EN">English</option>
              </select>
            </div>
          </div>

          {/* Content */}
          <div className="form-group">
            <label htmlFor="content">내용 *</label>
            <textarea
              id="content"
              name="content"
              value={formData.content}
              onChange={handleInputChange}
              placeholder="세탁 지식 내용을 입력하세요..."
              rows={10}
              required
            />
          </div>

          {/* File Upload */}
          <div className="form-group">
            <label htmlFor="files">파일 업로드</label>
            <div className="file-upload-area">
              <input
                type="file"
                id="files"
                multiple
                accept=".pdf,.txt,.md,.jpg,.jpeg,.png,.gif,.webp"
                onChange={handleFileSelect}
                style={{ display: 'none' }}
              />
              <button
                type="button"
                className="file-upload-button"
                onClick={() => document.getElementById('files')?.click()}
              >
                📁 파일 선택 (PDF, TXT, 이미지)
              </button>
              <p className="file-hint">
                PDF, 텍스트 파일, 이미지를 업로드하면 자동으로 텍스트가 추출됩니다
              </p>
            </div>

            {selectedFiles.length > 0 && (
              <div className="selected-files">
                {selectedFiles.map((file, index) => (
                  <div key={index} className="file-item">
                    <span className="file-icon">{getFileIcon(file.name)}</span>
                    <span className="file-name">{file.name}</span>
                    <span className="file-size">
                      ({(file.size / 1024).toFixed(1)} KB)
                    </span>
                    <button
                      type="button"
                      className="remove-file-button"
                      onClick={() => handleRemoveFile(index)}
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Tags */}
          <div className="form-group">
            <label htmlFor="tags">태그</label>
            <div className="tag-input-container">
              <input
                type="text"
                id="tag-input"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddTag())}
                placeholder="태그 입력 후 Enter 또는 추가 버튼 클릭"
              />
              <button type="button" onClick={handleAddTag} className="add-tag-button">
                추가
              </button>
            </div>

            {formData.tags.length > 0 && (
              <div className="tags-container">
                {formData.tags.map((tag, index) => (
                  <span key={index} className="tag">
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="remove-tag-button"
                    >
                      ✕
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Submit Button */}
          <div className="form-actions">
            <button type="submit" className="submit-button" disabled={isLoading}>
              {isLoading ? '등록 중...' : '지식 등록'}
            </button>
          </div>
        </form>

        {/* Knowledge List */}
        <div className="knowledge-list-section">
          <h2>등록된 지식 목록</h2>
          {isLoadingList ? (
            <p className="loading-text">불러오는 중...</p>
          ) : knowledgeList.length === 0 ? (
            <p className="empty-text">등록된 지식이 없습니다.</p>
          ) : (
            <div className="knowledge-list">
              {knowledgeList.map((item) => (
                <div key={item.id} className="knowledge-item">
                  <div className="knowledge-item-header">
                    <h3>{item.title}</h3>
                    <button
                      className="delete-button"
                      onClick={() => handleDelete(item.id)}
                    >
                      삭제
                    </button>
                  </div>
                  <div className="knowledge-item-meta">
                    <span className="category-badge">
                      {categories.find(c => c.value === item.category)?.label || item.category}
                    </span>
                    <span className="difficulty-badge">
                      {item.difficulty === 'basic' ? '기본' : '전문가'}
                    </span>
                    <span className="language-badge">{item.language}</span>
                  </div>
                  <p className="knowledge-item-content">
                    {item.content.length > 200 ? item.content.substring(0, 200) + '...' : item.content}
                  </p>
                  {item.tags && item.tags.length > 0 && (
                    <div className="knowledge-item-tags">
                      {item.tags.map((tag, idx) => (
                        <span key={idx} className="tag-badge">{tag}</span>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AdminKnowledgePage;
